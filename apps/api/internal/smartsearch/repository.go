package smartsearch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"fungi-wiki/apps/api/internal/embedding"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool       *pgxpool.Pool
	embeddings embedding.Provider
}

func NewRepository(pool *pgxpool.Pool, provider embedding.Provider) *Repository {
	return &Repository{pool: pool, embeddings: provider}
}

func (r *Repository) Search(ctx context.Context, p Params) (Result, error) {
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	terms := []string{}
	query := strings.TrimSpace(p.Query)
	if query != "" {
		terms = append(terms, query)
		rows, err := r.pool.Query(ctx, `SELECT CASE WHEN lower(term)=lower($1) OR lower($1) LIKE '%'||lower(term)||'%' THEN synonym ELSE term END FROM search_synonyms WHERE enabled AND (lower(term)=lower($1) OR lower($1) LIKE '%'||lower(term)||'%' OR lower(synonym)=lower($1) OR lower($1) LIKE '%'||lower(synonym)||'%') ORDER BY weight DESC`, query)
		if err != nil {
			return Result{}, err
		}
		for rows.Next() {
			var term string
			if err = rows.Scan(&term); err != nil {
				rows.Close()
				return Result{}, err
			}
			terms = append(terms, term)
		}
		rows.Close()
	}
	patterns := make([]string, len(terms))
	for i, term := range terms {
		patterns[i] = "%" + term + "%"
	}
	vector := []float32{}
	semantic := false
	if query != "" && r.embeddings.Enabled() {
		vectors, err := r.embeddings.Embed(ctx, []string{query})
		if err == nil && len(vectors) == 1 {
			vector = vectors[0]
			semantic = true
		}
	}
	const rankedSQL = `WITH ranked AS (
SELECT s.id::text,s.slug,s.latin_name,COALESCE(s.chinese_name,''),COALESCE(s.strain_number,''),COALESCE(s.source_environment,''),COALESCE(s.safety_level,''),s.is_model_organism,COALESCE(s.summary,''),s.status,s.data_quality_score,s.created_at,s.updated_at,s.published_at,
GREATEST(CASE WHEN $1='' THEN 0 ELSE similarity(s.slug,$1) END,CASE WHEN $1='' THEN 0 ELSE similarity(s.latin_name,$1) END,CASE WHEN $1='' THEN 0 ELSE similarity(COALESCE(s.chinese_name,''),$1) END,COALESCE((SELECT MAX(similarity(sa.alias_name,$1)) FROM species_aliases sa WHERE sa.species_id=s.id),0)) AS keyword_score,
COALESCE(cosine_similarity(se.embedding,$3::real[]),0) AS semantic_score,
COALESCE((SELECT MAX(rr.boost) FROM search_recall_rules rr WHERE rr.enabled AND lower($1) LIKE '%'||lower(rr.query_pattern)||'%' AND (rr.safety_level IS NULL OR rr.safety_level=s.safety_level) AND (rr.function_tag_code IS NULL OR EXISTS(SELECT 1 FROM species_functions rsf JOIN function_tags rft ON rft.id=rsf.function_tag_id WHERE rsf.species_id=s.id AND rft.code=rr.function_tag_code))),0) AS rule_boost
FROM species s LEFT JOIN species_embeddings se ON se.species_id=s.id
WHERE s.status='published'
AND ($4='' OR EXISTS(SELECT 1 FROM species_functions sf JOIN function_tags ft ON ft.id=sf.function_tag_id WHERE sf.species_id=s.id AND (ft.id::text=$4 OR ft.code=$4)))
AND ($5::double precision IS NULL OR EXISTS(SELECT 1 FROM culture_conditions cc WHERE cc.species_id=s.id AND (cc.temperature_min IS NULL OR cc.temperature_min<=$5) AND (cc.temperature_max IS NULL OR cc.temperature_max>=$5)))
AND ($6::double precision IS NULL OR EXISTS(SELECT 1 FROM culture_conditions cc WHERE cc.species_id=s.id AND (cc.ph_min IS NULL OR cc.ph_min<=$6) AND (cc.ph_max IS NULL OR cc.ph_max>=$6)))
AND ($7='' OR s.safety_level ILIKE $7) AND ($8='' OR s.source_environment ILIKE '%'||$8||'%')
AND ($1='' OR EXISTS(SELECT 1 FROM unnest($2::text[]) pattern WHERE s.slug ILIKE pattern OR s.latin_name ILIKE pattern OR COALESCE(s.chinese_name,'') ILIKE pattern OR COALESCE(s.summary,'') ILIKE pattern OR EXISTS(SELECT 1 FROM species_aliases sa WHERE sa.species_id=s.id AND sa.alias_name ILIKE pattern) OR EXISTS(SELECT 1 FROM species_functions sf JOIN function_tags ft ON ft.id=sf.function_tag_id WHERE sf.species_id=s.id AND (ft.name ILIKE pattern OR ft.description ILIKE pattern))) OR COALESCE(cosine_similarity(se.embedding,$3::real[]),0)>=0.35)
)`
	args := []any{query, patterns, vector, p.FunctionTag, p.Temperature, p.PH, p.SafetyLevel, p.SourceEnvironment}
	var total int
	if err := r.pool.QueryRow(ctx, rankedSQL+` SELECT COUNT(*) FROM ranked`, args...).Scan(&total); err != nil {
		return Result{}, err
	}
	orderBy := "hybrid_score DESC,updated_at DESC"
	switch p.Sort {
	case "name":
		orderBy = "latin_name ASC"
	case "quality":
		orderBy = "data_quality_score DESC,updated_at DESC"
	case "oldest":
		orderBy = "updated_at ASC"
	case "updated":
		orderBy = "updated_at DESC"
	}
	rows, err := r.pool.Query(ctx, rankedSQL+` SELECT *, (keyword_score*0.45+semantic_score*0.40+LEAST(data_quality_score/100.0,1)*0.10+rule_boost) AS hybrid_score FROM ranked ORDER BY `+orderBy+` LIMIT $9 OFFSET $10`, append(args, p.Limit, p.Offset)...)
	if err != nil {
		return Result{}, err
	}
	defer rows.Close()
	items := []ResultItem{}
	for rows.Next() {
		var item ResultItem
		var keywordScore, semanticScore, ruleBoost float64
		if err = rows.Scan(&item.ID, &item.Slug, &item.LatinName, &item.ChineseName, &item.StrainNumber, &item.SourceEnvironment, &item.SafetyLevel, &item.IsModelOrganism, &item.Summary, &item.Status, &item.DataQualityScore, &item.CreatedAt, &item.UpdatedAt, &item.PublishedAt, &keywordScore, &semanticScore, &ruleBoost, &item.HybridScore); err != nil {
			return Result{}, err
		}
		if keywordScore > 0.15 {
			item.MatchReasons = append(item.MatchReasons, "关键词匹配")
		}
		if semanticScore >= 0.35 {
			item.MatchReasons = append(item.MatchReasons, "语义相似")
		}
		if ruleBoost > 0 {
			item.MatchReasons = append(item.MatchReasons, "运营召回规则")
		}
		items = append(items, item)
	}
	return Result{Items: items, Total: total, Limit: p.Limit, Offset: p.Offset, SemanticEnabled: semantic, ExpandedTerms: terms}, rows.Err()
}

func (r *Repository) Reindex(ctx context.Context) (int, error) {
	if !r.embeddings.Enabled() {
		return 0, embedding.ErrDisabled
	}
	rows, err := r.pool.Query(ctx, `SELECT s.id::text,concat_ws(E'\n',s.latin_name,s.chinese_name,s.summary,s.source_environment,string_agg(DISTINCT ft.name||' '||COALESCE(ft.description,''),E'\n'),string_agg(DISTINCT COALESCE(sf.description,''),E'\n')) FROM species s LEFT JOIN species_functions sf ON sf.species_id=s.id LEFT JOIN function_tags ft ON ft.id=sf.function_tag_id WHERE s.status='published' GROUP BY s.id`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	ids, texts := []string{}, []string{}
	for rows.Next() {
		var id, text string
		if err = rows.Scan(&id, &text); err != nil {
			return 0, err
		}
		ids = append(ids, id)
		texts = append(texts, text)
	}
	if len(texts) == 0 {
		return 0, nil
	}
	vectors, err := r.embeddings.Embed(ctx, texts)
	if err != nil {
		return 0, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	for i, v := range vectors {
		hash := sha256.Sum256([]byte(texts[i]))
		_, err = tx.Exec(ctx, `INSERT INTO species_embeddings(species_id,model,dimensions,embedding,content_hash,embedded_at)VALUES($1::uuid,$2,$3,$4,$5,NOW()) ON CONFLICT(species_id)DO UPDATE SET model=EXCLUDED.model,dimensions=EXCLUDED.dimensions,embedding=EXCLUDED.embedding,content_hash=EXCLUDED.content_hash,embedded_at=NOW()`, ids[i], r.embeddings.Model(), len(v), v, hex.EncodeToString(hash[:]))
		if err != nil {
			return 0, err
		}
	}
	return len(vectors), tx.Commit(ctx)
}

func (r *Repository) LogHistory(ctx context.Context, userID string, p Params, count int) error {
	filters, _ := json.Marshal(map[string]any{"functionTag": p.FunctionTag, "temperature": p.Temperature, "ph": p.PH, "safetyLevel": p.SafetyLevel, "sourceEnvironment": p.SourceEnvironment})
	_, err := r.pool.Exec(ctx, `INSERT INTO user_search_history(user_id,query,filters,result_count)VALUES($1::uuid,$2,$3,$4)`, userID, p.Query, filters, count)
	return err
}
func (r *Repository) History(ctx context.Context, userID string) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,query,filters,result_count,created_at FROM user_search_history WHERE user_id=$1::uuid ORDER BY created_at DESC LIMIT 100`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, q string
		var filters map[string]any
		var count int
		var created any
		if err = rows.Scan(&id, &q, &filters, &count, &created); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "query": q, "filters": filters, "resultCount": count, "createdAt": created})
	}
	return items, rows.Err()
}
func (r *Repository) ClearHistory(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_search_history WHERE user_id=$1::uuid`, userID)
	return err
}
func (r *Repository) Favorite(ctx context.Context, userID, speciesID string, add bool) error {
	if add {
		_, err := r.pool.Exec(ctx, `INSERT INTO user_species_favorites(user_id,species_id) SELECT $1::uuid,id FROM species WHERE (id::text=$2 OR slug=$2) AND status='published' ON CONFLICT DO NOTHING`, userID, speciesID)
		return err
	}
	_, err := r.pool.Exec(ctx, `DELETE FROM user_species_favorites WHERE user_id=$1::uuid AND species_id IN(SELECT id FROM species WHERE id::text=$2 OR slug=$2)`, userID, speciesID)
	return err
}
func (r *Repository) Favorites(ctx context.Context, userID string) ([]ResultItem, error) {
	rows, err := r.pool.Query(ctx, `SELECT s.id::text,s.slug,s.latin_name,COALESCE(s.chinese_name,''),COALESCE(s.strain_number,''),COALESCE(s.source_environment,''),COALESCE(s.safety_level,''),s.is_model_organism,COALESCE(s.summary,''),s.status,s.data_quality_score,s.created_at,s.updated_at,s.published_at FROM user_species_favorites f JOIN species s ON s.id=f.species_id WHERE f.user_id=$1::uuid AND s.status='published' ORDER BY f.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ResultItem{}
	for rows.Next() {
		var x ResultItem
		if err = rows.Scan(&x.ID, &x.Slug, &x.LatinName, &x.ChineseName, &x.StrainNumber, &x.SourceEnvironment, &x.SafetyLevel, &x.IsModelOrganism, &x.Summary, &x.Status, &x.DataQualityScore, &x.CreatedAt, &x.UpdatedAt, &x.PublishedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

func (r *Repository) ListSynonyms(ctx context.Context) ([]Synonym, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,term,synonym,weight,enabled,created_at,updated_at FROM search_synonyms ORDER BY term,synonym`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Synonym{}
	for rows.Next() {
		var x Synonym
		if err = rows.Scan(&x.ID, &x.Term, &x.Value, &x.Weight, &x.Enabled, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) SaveSynonym(ctx context.Context, id string, in SynonymInput) (Synonym, error) {
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	if in.Weight <= 0 {
		in.Weight = .85
	}
	query := `INSERT INTO search_synonyms(term,synonym,weight,enabled)VALUES($1,$2,$3,$4)RETURNING id::text,term,synonym,weight,enabled,created_at,updated_at`
	args := []any{in.Term, in.Value, in.Weight, enabled}
	if id != "" {
		query = `UPDATE search_synonyms SET term=$2,synonym=$3,weight=$4,enabled=$5 WHERE id::text=$1 RETURNING id::text,term,synonym,weight,enabled,created_at,updated_at`
		args = []any{id, in.Term, in.Value, in.Weight, enabled}
	}
	var x Synonym
	err := r.pool.QueryRow(ctx, query, args...).Scan(&x.ID, &x.Term, &x.Value, &x.Weight, &x.Enabled, &x.CreatedAt, &x.UpdatedAt)
	return x, err
}
func (r *Repository) DeleteSynonym(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM search_synonyms WHERE id::text=$1`, id)
	return err
}
func (r *Repository) ListRules(ctx context.Context) ([]Rule, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text,name,query_pattern,COALESCE(function_tag_code,''),COALESCE(safety_level,''),boost,enabled,created_at,updated_at FROM search_recall_rules ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Rule{}
	for rows.Next() {
		var x Rule
		if err = rows.Scan(&x.ID, &x.Name, &x.QueryPattern, &x.FunctionTagCode, &x.SafetyLevel, &x.Boost, &x.Enabled, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) SaveRule(ctx context.Context, id string, in RuleInput) (Rule, error) {
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	if in.Boost == 0 {
		in.Boost = .15
	}
	query := `INSERT INTO search_recall_rules(name,query_pattern,function_tag_code,safety_level,boost,enabled)VALUES($1,$2,NULLIF($3,''),NULLIF($4,''),$5,$6)RETURNING id::text,name,query_pattern,COALESCE(function_tag_code,''),COALESCE(safety_level,''),boost,enabled,created_at,updated_at`
	args := []any{in.Name, in.QueryPattern, in.FunctionTagCode, in.SafetyLevel, in.Boost, enabled}
	if id != "" {
		query = `UPDATE search_recall_rules SET name=$2,query_pattern=$3,function_tag_code=NULLIF($4,''),safety_level=NULLIF($5,''),boost=$6,enabled=$7 WHERE id::text=$1 RETURNING id::text,name,query_pattern,COALESCE(function_tag_code,''),COALESCE(safety_level,''),boost,enabled,created_at,updated_at`
		args = []any{id, in.Name, in.QueryPattern, in.FunctionTagCode, in.SafetyLevel, in.Boost, enabled}
	}
	var x Rule
	err := r.pool.QueryRow(ctx, query, args...).Scan(&x.ID, &x.Name, &x.QueryPattern, &x.FunctionTagCode, &x.SafetyLevel, &x.Boost, &x.Enabled, &x.CreatedAt, &x.UpdatedAt)
	return x, err
}
func (r *Repository) DeleteRule(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM search_recall_rules WHERE id::text=$1`, id)
	return err
}
func parseFloat(value string) (*float64, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var x float64
	_, err := fmt.Sscan(value, &x)
	return &x, err
}
