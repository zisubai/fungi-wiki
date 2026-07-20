package recommendation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"math"
	"sort"
	"strings"
)

var ErrNotFound = errors.New("recommendation not found")

type Repository interface {
	Recommend(context.Context, Input) (Response, error)
	RecommendCombination(context.Context, CombinationInput) (CombinationResponse, error)
	Feedback(context.Context, string, FeedbackInput) error
	CombinationFeedback(context.Context, string, FeedbackInput) error
	CreateCombinationExperiment(context.Context, string, CombinationExperimentInput) (CombinationExperiment, error)
	Quality(context.Context, int) (QualityReport, error)
}

type combinationCandidate struct {
	member                         CombinationMember
	quality                        float64
	temperatureMin, temperatureMax *float64
	phMin, phMax                   *float64
}

func (r *PostgresRepository) RecommendCombination(ctx context.Context, input CombinationInput) (CombinationResponse, error) {
	response := CombinationResponse{Items: []Combination{}, Disclaimer: "组合建议仅用于候选设计，必须进行拮抗性、共培养稳定性、生物安全和功能验证实验。"}
	first, err := r.listCombinationCandidates(ctx, input.FunctionTags[0], input.SafetyLevel)
	if err != nil {
		return response, err
	}
	second, err := r.listCombinationCandidates(ctx, input.FunctionTags[1], input.SafetyLevel)
	if err != nil {
		return response, err
	}
	response.Items = rankCombinations(first, second, 5)
	inputJSON, _ := json.Marshal(input.FunctionTags)
	itemsJSON, _ := json.Marshal(response.Items)
	riskLevel := "low"
	for _, item := range response.Items {
		if !item.Compatible || item.Warning != "" {
			riskLevel = "review_required"
			break
		}
	}
	if len(response.Items) == 0 {
		riskLevel = "no_candidate"
	}
	err = r.pool.QueryRow(ctx, `INSERT INTO combination_recommendation_records(function_tags,safety_level,combinations,model_name,risk_level)VALUES($1::jsonb,NULLIF($2,''),$3::jsonb,'combination-rules-v2',$4)RETURNING id::text`, string(inputJSON), input.SafetyLevel, string(itemsJSON), riskLevel).Scan(&response.RecordID)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (r *PostgresRepository) listCombinationCandidates(ctx context.Context, tag, safetyLevel string) ([]combinationCandidate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.id::text,s.slug,s.latin_name,COALESCE(s.chinese_name,''),COALESCE(s.safety_level,''),
		       ARRAY[$1::text],COUNT(DISTINCT e.id),s.data_quality_score,
		       MIN(cc.temperature_min),MAX(cc.temperature_max),MIN(cc.ph_min),MAX(cc.ph_max)
		FROM species s
		JOIN species_functions sf ON sf.species_id=s.id
		JOIN function_tags ft ON ft.id=sf.function_tag_id
		LEFT JOIN culture_conditions cc ON cc.species_id=s.id
		LEFT JOIN evidences e ON e.species_id=s.id
		WHERE s.status='published' AND (ft.code=$1 OR ft.id::text=$1)
		  AND ($2='' OR s.safety_level ILIKE $2)
		GROUP BY s.id ORDER BY s.data_quality_score DESC,COUNT(DISTINCT e.id) DESC LIMIT 20
	`, tag, safetyLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	candidates := make([]combinationCandidate, 0)
	for rows.Next() {
		var candidate combinationCandidate
		if err = rows.Scan(
			&candidate.member.ID, &candidate.member.Slug, &candidate.member.LatinName, &candidate.member.ChineseName,
			&candidate.member.SafetyLevel, &candidate.member.FunctionTags, &candidate.member.EvidenceCount, &candidate.quality,
			&candidate.temperatureMin, &candidate.temperatureMax, &candidate.phMin, &candidate.phMax,
		); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

func rankCombinations(first, second []combinationCandidate, limit int) []Combination {
	items := make([]Combination, 0)
	seen := make(map[string]struct{})
	for _, left := range first {
		for _, right := range second {
			if left.member.ID == right.member.ID {
				continue
			}
			ids := []string{left.member.ID, right.member.ID}
			sort.Strings(ids)
			key := strings.Join(ids, ":")
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			items = append(items, buildCombination(left, right))
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Compatible != items[j].Compatible {
			return items[i].Compatible
		}
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		return combinationEvidenceCount(items[i]) > combinationEvidenceCount(items[j])
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func combinationEvidenceCount(item Combination) int {
	total := 0
	for _, member := range item.Members {
		total += member.EvidenceCount
	}
	return total
}

func buildCombination(first, second combinationCandidate) Combination {
	temperatureMin, temperatureMax, temperatureOK := overlap(first.temperatureMin, first.temperatureMax, second.temperatureMin, second.temperatureMax)
	phMin, phMax, phOK := overlap(first.phMin, first.phMax, second.phMin, second.phMax)
	compatible := temperatureOK && phOK
	combination := Combination{
		Members: []CombinationMember{first.member, second.member}, Score: math.Round((first.quality+second.quality)/2*10) / 10,
		TemperatureMin: temperatureMin, TemperatureMax: temperatureMax, PHMin: phMin, PHMax: phMax, Compatible: compatible,
		Reasons: []string{"两个候选菌种覆盖所选的两个功能", fmt.Sprintf("共关联 %d 条文献证据", first.member.EvidenceCount+second.member.EvidenceCount)},
	}
	if compatible {
		combination.Reasons = append(combination.Reasons, "现有培养数据存在温度和 pH 交集")
	} else {
		combination.Warning = "现有数据无法证明存在共同培养窗口，需单独培养或实验优化。"
	}
	for _, member := range combination.Members {
		if member.SafetyLevel != "" && !strings.EqualFold(member.SafetyLevel, "BSL-1") {
			combination.Warning = "组合包含非 BSL-1 菌种，必须完成专业生物安全评估。"
		}
	}
	return combination
}

func overlap(firstMin, firstMax, secondMin, secondMax *float64) (*float64, *float64, bool) {
	if firstMin == nil || firstMax == nil || secondMin == nil || secondMax == nil {
		return nil, nil, false
	}
	minimum, maximum := math.Max(*firstMin, *secondMin), math.Min(*firstMax, *secondMax)
	return &minimum, &maximum, minimum <= maximum
}

func (r *PostgresRepository) Feedback(ctx context.Context, recordID string, input FeedbackInput) error {
	command, err := r.pool.Exec(ctx, `INSERT INTO user_feedback(entity_type,entity_id,feedback_type,content,status)SELECT 'recommendation',id,$2,NULLIF($3,''),'open' FROM recommendation_records WHERE id=$1::uuid`, recordID, input.FeedbackType, strings.TrimSpace(input.Content))
	if err == nil && command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r *PostgresRepository) CombinationFeedback(ctx context.Context, recordID string, input FeedbackInput) error {
	command, err := r.pool.Exec(ctx, `INSERT INTO user_feedback(entity_type,entity_id,feedback_type,content,status)SELECT 'combination_recommendation',id,$2,NULLIF($3,''),'open' FROM combination_recommendation_records WHERE id=$1::uuid`, recordID, input.FeedbackType, strings.TrimSpace(input.Content))
	if err == nil && command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r *PostgresRepository) CreateCombinationExperiment(ctx context.Context, recordID string, input CombinationExperimentInput) (CombinationExperiment, error) {
	var experiment CombinationExperiment
	var membersJSON []byte
	err := r.pool.QueryRow(ctx, `INSERT INTO combination_experiments(combination_record_id,candidate_index,candidate_members,outcome,temperature,ph,notes)
		SELECT id,$2::int,combinations->($2::int)->'members',$3,$4,$5,$6 FROM combination_recommendation_records
		WHERE id=$1::uuid AND jsonb_array_length(combinations)>$2::int
		RETURNING id::text,combination_record_id::text,candidate_index,candidate_members,outcome,temperature,ph,notes,created_at`,
		recordID, *input.CandidateIndex, input.Outcome, input.Temperature, input.PH, strings.TrimSpace(input.Notes),
	).Scan(&experiment.ID, &experiment.CombinationRecordID, &experiment.CandidateIndex, &membersJSON, &experiment.Outcome, &experiment.Temperature, &experiment.PH, &experiment.Notes, &experiment.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return experiment, ErrNotFound
	}
	_ = json.Unmarshal(membersJSON, &experiment.CandidateMembers)
	return experiment, err
}

func (r *PostgresRepository) Quality(ctx context.Context, limit int) (QualityReport, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	report := QualityReport{Records: []QualityRecord{}, Combinations: []CombinationQualityRecord{}}
	err := r.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT r.id),COUNT(f.id)FILTER(WHERE feedback_type='helpful'),COUNT(f.id)FILTER(WHERE feedback_type='unhelpful') FROM recommendation_records r LEFT JOIN user_feedback f ON f.entity_type='recommendation' AND f.entity_id=r.id`).Scan(&report.Total, &report.Helpful, &report.Unhelpful)
	if err != nil {
		return report, err
	}
	rows, err := r.pool.Query(ctx, `SELECT r.id::text,r.requirement,r.parsed_intent,r.recommended_species,COALESCE(r.model_name,''),COALESCE(r.risk_level,''),COUNT(f.id)FILTER(WHERE f.feedback_type='helpful'),COUNT(f.id)FILTER(WHERE f.feedback_type='unhelpful'),r.created_at FROM recommendation_records r LEFT JOIN user_feedback f ON f.entity_type='recommendation' AND f.entity_id=r.id GROUP BY r.id ORDER BY r.created_at DESC LIMIT $1`, limit)
	if err != nil {
		return report, err
	}
	defer rows.Close()
	for rows.Next() {
		var record QualityRecord
		var intentJSON, itemsJSON []byte
		if err = rows.Scan(&record.ID, &record.Requirement, &intentJSON, &itemsJSON, &record.ModelName, &record.RiskLevel, &record.HelpfulCount, &record.UnhelpfulCount, &record.CreatedAt); err != nil {
			return report, err
		}
		_ = json.Unmarshal(intentJSON, &record.ParsedIntent)
		_ = json.Unmarshal(itemsJSON, &record.Items)
		report.Records = append(report.Records, record)
	}
	if err = rows.Err(); err != nil {
		return report, err
	}
	combinationRows, err := r.pool.Query(ctx, `SELECT c.id::text,c.function_tags,COALESCE(c.safety_level,''),c.combinations,c.model_name,c.risk_level,COUNT(f.id)FILTER(WHERE f.feedback_type='helpful'),COUNT(f.id)FILTER(WHERE f.feedback_type='unhelpful'),c.created_at FROM combination_recommendation_records c LEFT JOIN user_feedback f ON f.entity_type='combination_recommendation' AND f.entity_id=c.id GROUP BY c.id ORDER BY c.created_at DESC LIMIT $1`, limit)
	if err != nil {
		return report, err
	}
	defer combinationRows.Close()
	for combinationRows.Next() {
		var record CombinationQualityRecord
		var tagsJSON, itemsJSON []byte
		if err = combinationRows.Scan(&record.ID, &tagsJSON, &record.SafetyLevel, &itemsJSON, &record.ModelName, &record.RiskLevel, &record.HelpfulCount, &record.UnhelpfulCount, &record.CreatedAt); err != nil {
			return report, err
		}
		_ = json.Unmarshal(tagsJSON, &record.FunctionTags)
		_ = json.Unmarshal(itemsJSON, &record.Items)
		report.Combinations = append(report.Combinations, record)
	}
	if err = combinationRows.Err(); err != nil {
		return report, err
	}
	if len(report.Combinations) == 0 {
		return report, nil
	}
	indexes := make(map[string]int, len(report.Combinations))
	for index := range report.Combinations {
		report.Combinations[index].Experiments = []CombinationExperiment{}
		indexes[report.Combinations[index].ID] = index
	}
	experimentRows, err := r.pool.Query(ctx, `SELECT id::text,combination_record_id::text,candidate_index,candidate_members,outcome,temperature,ph,notes,created_at FROM combination_experiments ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return report, err
	}
	defer experimentRows.Close()
	for experimentRows.Next() {
		var experiment CombinationExperiment
		var membersJSON []byte
		if err = experimentRows.Scan(&experiment.ID, &experiment.CombinationRecordID, &experiment.CandidateIndex, &membersJSON, &experiment.Outcome, &experiment.Temperature, &experiment.PH, &experiment.Notes, &experiment.CreatedAt); err != nil {
			return report, err
		}
		_ = json.Unmarshal(membersJSON, &experiment.CandidateMembers)
		if index, ok := indexes[experiment.CombinationRecordID]; ok {
			report.Combinations[index].Experiments = append(report.Combinations[index].Experiments, experiment)
		}
	}
	return report, experimentRows.Err()
}

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(p *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{p} }

type candidate struct {
	item               Item
	quality            float64
	evidenceAverage    float64
	functionConfidence float64
	functionName       string
}

func (r *PostgresRepository) Recommend(ctx context.Context, in Input) (Response, error) {
	in = mergeParsedRequirement(in, parseRequirement(in.Requirement))
	limit := in.Limit
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	tag := strings.TrimSpace(in.FunctionTag)
	if tag == "" {
		_ = r.pool.QueryRow(ctx, `SELECT code FROM function_tags WHERE POSITION(LOWER(name) IN LOWER($1))>0 OR POSITION(LOWER(code) IN LOWER($1))>0 ORDER BY LENGTH(name) DESC LIMIT 1`, in.Requirement).Scan(&tag)
	}
	where := []string{"s.status='published'"}
	args := []any{}
	add := func(value any, condition string) {
		args = append(args, value)
		where = append(where, strings.ReplaceAll(condition, "%d", fmt.Sprint(len(args))))
	}
	if tag != "" {
		add(tag, `EXISTS(SELECT 1 FROM species_functions sf JOIN function_tags ft ON ft.id=sf.function_tag_id WHERE sf.species_id=s.id AND (ft.code=$%d OR ft.id::text=$%d))`)
	}
	if in.SafetyLevel != "" {
		add(in.SafetyLevel, `s.safety_level ILIKE $%d`)
	}
	if in.SourceEnvironment != "" {
		add("%"+in.SourceEnvironment+"%", `s.source_environment ILIKE $%d`)
	}
	if in.Temperature != nil {
		add(*in.Temperature, `EXISTS(SELECT 1 FROM culture_conditions cc WHERE cc.species_id=s.id AND (cc.temperature_min IS NULL OR cc.temperature_min<=$%d) AND (cc.temperature_max IS NULL OR cc.temperature_max>=$%d) AND (cc.temperature_min IS NOT NULL OR cc.temperature_max IS NOT NULL))`)
	}
	if in.PH != nil {
		add(*in.PH, `EXISTS(SELECT 1 FROM culture_conditions cc WHERE cc.species_id=s.id AND (cc.ph_min IS NULL OR cc.ph_min<=$%d) AND (cc.ph_max IS NULL OR cc.ph_max>=$%d) AND (cc.ph_min IS NOT NULL OR cc.ph_max IS NOT NULL))`)
	}
	args = append(args, tag)
	tagIndex := len(args)
	query := fmt.Sprintf(`SELECT s.id::text,s.slug,s.latin_name,COALESCE(s.chinese_name,''),COALESCE(s.safety_level,''),COALESCE(s.summary,''),s.data_quality_score,(SELECT COUNT(*) FROM evidences e WHERE e.species_id=s.id),(SELECT COALESCE(AVG(e.evidence_score),0) FROM evidences e WHERE e.species_id=s.id),(SELECT COALESCE(MAX(sf.confidence_score),0) FROM species_functions sf JOIN function_tags ft ON ft.id=sf.function_tag_id WHERE sf.species_id=s.id AND (ft.code=$%d OR ft.id::text=$%d)),COALESCE((SELECT ft.name FROM species_functions sf JOIN function_tags ft ON ft.id=sf.function_tag_id WHERE sf.species_id=s.id AND (ft.code=$%d OR ft.id::text=$%d) LIMIT 1),'') FROM species s WHERE %s LIMIT 100`, tagIndex, tagIndex, tagIndex, tagIndex, strings.Join(where, " AND "))
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return Response{}, err
	}
	defer rows.Close()
	candidates := []candidate{}
	for rows.Next() {
		var c candidate
		if err = rows.Scan(&c.item.ID, &c.item.Slug, &c.item.LatinName, &c.item.ChineseName, &c.item.SafetyLevel, &c.item.Summary, &c.quality, &c.item.EvidenceCount, &c.evidenceAverage, &c.functionConfidence, &c.functionName); err != nil {
			return Response{}, err
		}
		c.score(in, tag)
		candidates = append(candidates, c)
	}
	if err = rows.Err(); err != nil {
		return Response{}, err
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].item.Score > candidates[j].item.Score })
	items := []Item{}
	for i := 0; i < len(candidates) && i < limit; i++ {
		items = append(items, candidates[i].item)
	}
	if err = r.loadEvidenceReferences(ctx, items); err != nil {
		return Response{}, err
	}
	suggestions := []string{}
	if len(items) == 0 {
		suggestions, err = r.diagnoseNoResults(ctx, in, tag)
		if err != nil {
			return Response{}, err
		}
	}
	parsedIntent := map[string]any{"functionTag": tag, "temperature": in.Temperature, "ph": in.PH, "safetyLevel": in.SafetyLevel, "sourceEnvironment": in.SourceEnvironment}
	intent, _ := json.Marshal(parsedIntent)
	recommended, _ := json.Marshal(items)
	evidenceRefs := make([]EvidenceReference, 0)
	for _, item := range items {
		evidenceRefs = append(evidenceRefs, item.EvidenceReferences...)
	}
	evidenceJSON, _ := json.Marshal(evidenceRefs)
	risk := "low"
	for _, item := range items {
		if item.SafetyLevel != "" && !strings.EqualFold(item.SafetyLevel, "BSL-1") {
			risk = "review_required"
			break
		}
	}
	var recordID string
	err = r.pool.QueryRow(ctx, `INSERT INTO recommendation_records(requirement,parsed_intent,recommended_species,evidence_refs,model_name,risk_level)VALUES($1,$2::jsonb,$3::jsonb,$4::jsonb,'rules-v1',$5)RETURNING id::text`, in.Requirement, string(intent), string(recommended), string(evidenceJSON), risk).Scan(&recordID)
	if err != nil {
		return Response{}, err
	}
	return Response{RecordID: recordID, ParsedFunctionTag: tag, ParsedIntent: parsedIntent, Items: items, RelaxationSuggestions: suggestions, Disclaimer: "推荐仅用于候选初筛，不替代生物安全评估、专家判断和实验验证。"}, nil
}

type diagnosticCounts struct {
	total, function, safety, source, temperature, ph int
}

func (r *PostgresRepository) diagnoseNoResults(ctx context.Context, in Input, tag string) ([]string, error) {
	var temperature, ph any
	if in.Temperature != nil {
		temperature = *in.Temperature
	}
	if in.PH != nil {
		ph = *in.PH
	}
	var counts diagnosticCounts
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*),
		       COUNT(*) FILTER (WHERE $1 = '' OR EXISTS (SELECT 1 FROM species_functions sf JOIN function_tags ft ON ft.id=sf.function_tag_id WHERE sf.species_id=s.id AND (ft.code=$1 OR ft.id::text=$1))),
		       COUNT(*) FILTER (WHERE $2 = '' OR s.safety_level ILIKE $2),
		       COUNT(*) FILTER (WHERE $3 = '' OR s.source_environment ILIKE '%' || $3 || '%'),
		       COUNT(*) FILTER (WHERE $4::numeric IS NULL OR EXISTS (SELECT 1 FROM culture_conditions cc WHERE cc.species_id=s.id AND (cc.temperature_min IS NULL OR cc.temperature_min <= $4::numeric) AND (cc.temperature_max IS NULL OR cc.temperature_max >= $4::numeric) AND (cc.temperature_min IS NOT NULL OR cc.temperature_max IS NOT NULL))),
		       COUNT(*) FILTER (WHERE $5::numeric IS NULL OR EXISTS (SELECT 1 FROM culture_conditions cc WHERE cc.species_id=s.id AND (cc.ph_min IS NULL OR cc.ph_min <= $5::numeric) AND (cc.ph_max IS NULL OR cc.ph_max >= $5::numeric) AND (cc.ph_min IS NOT NULL OR cc.ph_max IS NOT NULL)))
		FROM species s WHERE s.status='published'
	`, tag, in.SafetyLevel, in.SourceEnvironment, temperature, ph).Scan(&counts.total, &counts.function, &counts.safety, &counts.source, &counts.temperature, &counts.ph)
	if err != nil {
		return nil, err
	}
	return buildRelaxationSuggestions(in, tag, counts), nil
}

func buildRelaxationSuggestions(in Input, tag string, counts diagnosticCounts) []string {
	if counts.total == 0 {
		return []string{"当前没有已发布菌种，请先在运营端完成数据发布。"}
	}
	suggestions := make([]string, 0, 5)
	if tag != "" && counts.function == 0 {
		suggestions = append(suggestions, "暂无菌种关联该功能，可取消功能限制或先补充功能标签数据。")
	}
	if in.SafetyLevel != "" && counts.safety == 0 {
		suggestions = append(suggestions, "暂无匹配安全等级的菌种，可放宽安全等级后再进行专业评估。")
	}
	if in.SourceEnvironment != "" && counts.source == 0 {
		suggestions = append(suggestions, "暂无匹配来源环境的菌种，可取消来源限制。")
	}
	if in.Temperature != nil && counts.temperature == 0 {
		suggestions = append(suggestions, "暂无培养温度覆盖目标值的菌种，可调整温度或补充培养条件。")
	}
	if in.PH != nil && counts.ph == 0 {
		suggestions = append(suggestions, "暂无培养 pH 覆盖目标值的菌种，可调整 pH 或补充培养条件。")
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "各条件单独均有候选，但组合后无交集；建议逐项取消条件定位冲突。")
	}
	return suggestions
}

func (r *PostgresRepository) loadEvidenceReferences(ctx context.Context, items []Item) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]string, 0, len(items))
	itemByID := make(map[string]*Item, len(items))
	for index := range items {
		ids = append(ids, items[index].ID)
		itemByID[items[index].ID] = &items[index]
		items[index].EvidenceReferences = []EvidenceReference{}
	}
	rows, err := r.pool.Query(ctx, `
		SELECT e.species_id::text, e.id::text, COALESCE(l.title, '未关联文献'), l.publication_year,
		       COALESCE(l.doi, ''), COALESCE(l.pmid, ''), COALESCE(l.source_url, ''),
		       e.conclusion, e.evidence_level, e.evidence_score
		FROM evidences e
		LEFT JOIN literatures l ON l.id = e.literature_id
		WHERE e.species_id::text = ANY($1)
		ORDER BY e.species_id, e.evidence_score DESC, e.created_at DESC
	`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var speciesID string
		var ref EvidenceReference
		if err = rows.Scan(&speciesID, &ref.ID, &ref.Title, &ref.PublicationYear, &ref.DOI, &ref.PMID, &ref.SourceURL, &ref.Conclusion, &ref.EvidenceLevel, &ref.EvidenceScore); err != nil {
			return err
		}
		item := itemByID[speciesID]
		if item != nil && len(item.EvidenceReferences) < 3 {
			item.EvidenceReferences = append(item.EvidenceReferences, ref)
		}
	}
	return rows.Err()
}
func (c *candidate) score(in Input, tag string) {
	score := math.Min(c.quality*.35, 35) + math.Min(float64(c.item.EvidenceCount)*6, 24) + math.Min(c.evidenceAverage*.15, 15)
	if tag != "" {
		score += 20 + math.Min(c.functionConfidence*.06, 6)
		c.item.Reasons = append(c.item.Reasons, "匹配功能："+c.functionName)
	}
	if in.Temperature != nil {
		c.item.Reasons = append(c.item.Reasons, fmt.Sprintf("培养温度范围覆盖 %.1f°C", *in.Temperature))
	}
	if in.PH != nil {
		c.item.Reasons = append(c.item.Reasons, fmt.Sprintf("培养 pH 范围覆盖 %.1f", *in.PH))
	}
	if c.item.EvidenceCount > 0 {
		c.item.Reasons = append(c.item.Reasons, fmt.Sprintf("关联 %d 条文献证据", c.item.EvidenceCount))
	} else {
		c.item.Reasons = append(c.item.Reasons, "暂无结构化文献证据，置信度较低")
	}
	if c.quality > 0 {
		c.item.Reasons = append(c.item.Reasons, fmt.Sprintf("数据质量评分 %.0f", c.quality))
	}
	if c.item.SafetyLevel != "" && !strings.EqualFold(c.item.SafetyLevel, "BSL-1") {
		c.item.RiskWarning = "该菌种不是 BSL-1，使用前必须完成专业生物安全评估。"
	}
	c.item.Score = math.Round(math.Min(score, 100)*10) / 10
}
