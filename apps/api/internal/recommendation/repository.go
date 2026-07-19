package recommendation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"math"
	"sort"
	"strings"
)

var ErrNotFound = errors.New("recommendation not found")

type Repository interface {
	Recommend(context.Context, Input) (Response, error)
	Feedback(context.Context, string, FeedbackInput) error
	Quality(context.Context, int) (QualityReport, error)
}

func (r *PostgresRepository) Feedback(ctx context.Context, recordID string, input FeedbackInput) error {
	command, err := r.pool.Exec(ctx, `INSERT INTO user_feedback(entity_type,entity_id,feedback_type,content,status)SELECT 'recommendation',id,$2,NULLIF($3,''),'open' FROM recommendation_records WHERE id=$1::uuid`, recordID, input.FeedbackType, strings.TrimSpace(input.Content))
	if err == nil && command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r *PostgresRepository) Quality(ctx context.Context, limit int) (QualityReport, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	report := QualityReport{Records: []QualityRecord{}}
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
	return report, rows.Err()
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
	intent, _ := json.Marshal(map[string]any{"functionTag": tag, "temperature": in.Temperature, "ph": in.PH, "safetyLevel": in.SafetyLevel, "sourceEnvironment": in.SourceEnvironment})
	recommended, _ := json.Marshal(items)
	risk := "low"
	for _, item := range items {
		if item.SafetyLevel != "" && !strings.EqualFold(item.SafetyLevel, "BSL-1") {
			risk = "review_required"
			break
		}
	}
	var recordID string
	err = r.pool.QueryRow(ctx, `INSERT INTO recommendation_records(requirement,parsed_intent,recommended_species,evidence_refs,model_name,risk_level)VALUES($1,$2::jsonb,$3::jsonb,'[]'::jsonb,'rules-v1',$4)RETURNING id::text`, in.Requirement, string(intent), string(recommended), risk).Scan(&recordID)
	if err != nil {
		return Response{}, err
	}
	return Response{RecordID: recordID, ParsedFunctionTag: tag, Items: items, Disclaimer: "推荐仅用于候选初筛，不替代生物安全评估、专家判断和实验验证。"}, nil
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
