package species

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("species not found")
var ErrDuplicateSlug = errors.New("species slug already exists")

type Repository interface {
	List(ctx context.Context, params ListParams) (ListResult, error)
	Compare(ctx context.Context, ids []string) ([]Comparison, error)
	Quality(ctx context.Context, idOrSlug string) (QualityReport, error)
	Get(ctx context.Context, idOrSlug string) (Species, error)
	Create(ctx context.Context, input CreateInput) (Species, error)
	Update(ctx context.Context, idOrSlug string, input UpdateInput) (Species, error)
	Archive(ctx context.Context, idOrSlug string) error
	Delete(ctx context.Context, idOrSlug string) error
}

func (repo *PostgresRepository) Quality(ctx context.Context, idOrSlug string) (QualityReport, error) {
	var score float64
	completed := make([]bool, 10)
	err := repo.pool.QueryRow(ctx, `
		SELECT data_quality_score,
		       NULLIF(BTRIM(latin_name), '') IS NOT NULL,
		       NULLIF(BTRIM(chinese_name), '') IS NOT NULL,
		       NULLIF(BTRIM(strain_number), '') IS NOT NULL,
		       NULLIF(BTRIM(source_environment), '') IS NOT NULL,
		       NULLIF(BTRIM(safety_level), '') IS NOT NULL,
		       NULLIF(BTRIM(summary), '') IS NOT NULL,
		       EXISTS(SELECT 1 FROM species_aliases sa WHERE sa.species_id=species.id),
		       EXISTS(SELECT 1 FROM species_functions sf WHERE sf.species_id=species.id),
		       EXISTS(SELECT 1 FROM culture_conditions cc WHERE cc.species_id=species.id),
		       EXISTS(SELECT 1 FROM evidences e WHERE e.species_id=species.id)
		FROM species WHERE id::text=$1 OR slug=$1 LIMIT 1
	`, idOrSlug).Scan(&score, &completed[0], &completed[1], &completed[2], &completed[3], &completed[4], &completed[5], &completed[6], &completed[7], &completed[8], &completed[9])
	if errors.Is(err, pgx.ErrNoRows) {
		return QualityReport{}, ErrNotFound
	}
	if err != nil {
		return QualityReport{}, err
	}
	return buildQualityReport(score, completed), nil
}

func buildQualityReport(score float64, completed []bool) QualityReport {
	definitions := []struct {
		key, label string
		weight     int
	}{
		{"latinName", "拉丁学名", 10}, {"chineseName", "中文名", 5}, {"strainNumber", "保藏/菌株编号", 5},
		{"sourceEnvironment", "来源环境", 10}, {"safetyLevel", "安全等级", 10}, {"summary", "菌种摘要", 15},
		{"aliases", "别名与同义词", 5}, {"functions", "功能标签关联", 15}, {"cultureConditions", "培养条件", 10},
		{"evidences", "文献证据", 15},
	}
	components := make([]QualityComponent, 0, len(definitions))
	for index, definition := range definitions {
		components = append(components, QualityComponent{Key: definition.key, Label: definition.label, Weight: definition.weight, Completed: index < len(completed) && completed[index]})
	}
	return QualityReport{Score: score, Components: components}
}

func (repo *PostgresRepository) Compare(ctx context.Context, ids []string) ([]Comparison, error) {
	rows, err := repo.pool.Query(ctx, `
		SELECT s.id::text, s.slug, s.latin_name, COALESCE(s.chinese_name, ''), COALESCE(s.strain_number, ''),
		       COALESCE(s.source_environment, ''), COALESCE(s.safety_level, ''), s.is_model_organism,
		       COALESCE(s.summary, ''), s.status, s.data_quality_score, s.created_at, s.updated_at, s.published_at,
		       COALESCE(array_agg(DISTINCT ft.name) FILTER (WHERE ft.name IS NOT NULL), '{}'),
		       MIN(cc.temperature_min), MAX(cc.temperature_max), MIN(cc.ph_min), MAX(cc.ph_max),
		       COUNT(DISTINCT e.id)
		FROM species s
		LEFT JOIN species_functions sf ON sf.species_id = s.id
		LEFT JOIN function_tags ft ON ft.id = sf.function_tag_id
		LEFT JOIN culture_conditions cc ON cc.species_id = s.id
		LEFT JOIN evidences e ON e.species_id = s.id
		WHERE s.status = 'published' AND (s.id::text = ANY($1) OR s.slug = ANY($1))
		GROUP BY s.id
		ORDER BY array_position($1::text[], s.slug), s.latin_name
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Comparison, 0, len(ids))
	for rows.Next() {
		var item Comparison
		err = rows.Scan(
			&item.ID, &item.Slug, &item.LatinName, &item.ChineseName, &item.StrainNumber,
			&item.SourceEnvironment, &item.SafetyLevel, &item.IsModelOrganism, &item.Summary,
			&item.Status, &item.DataQualityScore, &item.CreatedAt, &item.UpdatedAt, &item.PublishedAt,
			&item.FunctionTags, &item.TemperatureMin, &item.TemperatureMax, &item.PHMin, &item.PHMax,
			&item.EvidenceCount,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repo *PostgresRepository) List(ctx context.Context, params ListParams) (ListResult, error) {
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	where := []string{"1 = 1"}
	args := make([]any, 0)
	searchQueryIndex := 0

	if params.Query != "" {
		query := strings.TrimSpace(params.Query)
		args = append(args, "%"+query+"%")
		patternIndex := len(args)
		args = append(args, query)
		searchQueryIndex = len(args)
		where = append(where, fmt.Sprintf(`(
			slug ILIKE $%d OR latin_name ILIKE $%d OR chinese_name ILIKE $%d OR summary ILIKE $%d
			OR similarity(slug, $%d) >= 0.25 OR similarity(latin_name, $%d) >= 0.25
			OR similarity(COALESCE(chinese_name, ''), $%d) >= 0.25
			OR EXISTS (SELECT 1 FROM species_aliases sa WHERE sa.species_id=species.id AND (sa.alias_name ILIKE $%d OR similarity(sa.alias_name, $%d) >= 0.25))
		)`, patternIndex, patternIndex, patternIndex, patternIndex, searchQueryIndex, searchQueryIndex, searchQueryIndex, patternIndex, searchQueryIndex))
	}

	if params.Status != "" {
		args = append(args, params.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}

	if params.FunctionTag != "" {
		args = append(args, params.FunctionTag)
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM species_functions sf
			JOIN function_tags ft ON ft.id = sf.function_tag_id
			WHERE sf.species_id = species.id AND (ft.id::text = $%d OR ft.code = $%d)
		)`, len(args), len(args)))
	}

	if params.Temperature != nil {
		args = append(args, *params.Temperature)
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM culture_conditions cc WHERE cc.species_id = species.id
			AND (cc.temperature_min IS NOT NULL OR cc.temperature_max IS NOT NULL)
			AND (cc.temperature_min IS NULL OR cc.temperature_min <= $%d)
			AND (cc.temperature_max IS NULL OR cc.temperature_max >= $%d)
		)`, len(args), len(args)))
	}

	if params.PH != nil {
		args = append(args, *params.PH)
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM culture_conditions cc WHERE cc.species_id = species.id
			AND (cc.ph_min IS NOT NULL OR cc.ph_max IS NOT NULL)
			AND (cc.ph_min IS NULL OR cc.ph_min <= $%d)
			AND (cc.ph_max IS NULL OR cc.ph_max >= $%d)
		)`, len(args), len(args)))
	}

	if params.SafetyLevel != "" {
		args = append(args, strings.TrimSpace(params.SafetyLevel))
		where = append(where, fmt.Sprintf("safety_level ILIKE $%d", len(args)))
	}

	if params.SourceEnvironment != "" {
		args = append(args, "%"+strings.TrimSpace(params.SourceEnvironment)+"%")
		where = append(where, fmt.Sprintf("source_environment ILIKE $%d", len(args)))
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := repo.pool.QueryRow(ctx, `SELECT COUNT(*) FROM species WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return ListResult{}, err
	}
	orderBy := "updated_at DESC"
	switch params.Sort {
	case "relevance":
		if searchQueryIndex > 0 {
			orderBy = fmt.Sprintf(`GREATEST(
				similarity(slug, $%d), similarity(latin_name, $%d),
				similarity(COALESCE(chinese_name, ''), $%d),
				COALESCE((SELECT MAX(similarity(sa.alias_name, $%d)) FROM species_aliases sa WHERE sa.species_id=species.id), 0)
			) DESC, data_quality_score DESC`, searchQueryIndex, searchQueryIndex, searchQueryIndex, searchQueryIndex)
		}
	case "name":
		orderBy = "latin_name ASC"
	case "quality":
		orderBy = "data_quality_score DESC, updated_at DESC"
	case "oldest":
		orderBy = "updated_at ASC"
	}
	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id::text, slug, latin_name, COALESCE(chinese_name, ''), COALESCE(strain_number, ''),
		       COALESCE(source_environment, ''), COALESCE(safety_level, ''), is_model_organism,
		       COALESCE(summary, ''), status, data_quality_score, created_at, updated_at, published_at
		FROM species
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, orderBy, len(args)-1, len(args))

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return ListResult{}, err
	}
	defer rows.Close()

	items := make([]Species, 0)
	for rows.Next() {
		item, err := scanSpecies(rows)
		if err != nil {
			return ListResult{}, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Total: total}, nil
}

func (repo *PostgresRepository) LogSearch(ctx context.Context, params ListParams, resultCount int) error {
	filters, _ := json.Marshal(map[string]any{"functionTag": params.FunctionTag, "temperature": params.Temperature, "ph": params.PH, "safetyLevel": params.SafetyLevel, "sourceEnvironment": params.SourceEnvironment, "sort": params.Sort})
	_, err := repo.pool.Exec(ctx, `INSERT INTO search_logs(query,filters,result_count)VALUES($1,$2::jsonb,$3)`, strings.TrimSpace(params.Query), string(filters), resultCount)
	return err
}

func (repo *PostgresRepository) Get(ctx context.Context, idOrSlug string) (Species, error) {
	query := `
		SELECT id::text, slug, latin_name, COALESCE(chinese_name, ''), COALESCE(strain_number, ''),
		       COALESCE(source_environment, ''), COALESCE(safety_level, ''), is_model_organism,
		       COALESCE(summary, ''), status, data_quality_score, created_at, updated_at, published_at
		FROM species
		WHERE id::text = $1 OR slug = $1
		LIMIT 1
	`

	item, err := scanSpecies(repo.pool.QueryRow(ctx, query, idOrSlug))
	if errors.Is(err, pgx.ErrNoRows) {
		return Species{}, ErrNotFound
	}
	return item, err
}

func (repo *PostgresRepository) Create(ctx context.Context, input CreateInput) (Species, error) {
	query := `
		INSERT INTO species (slug, latin_name, chinese_name, strain_number, source_environment, safety_level, is_model_organism, summary, status, published_at)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7, NULLIF($8, ''), 'draft', NULL)
		RETURNING id::text, slug, latin_name, COALESCE(chinese_name, ''), COALESCE(strain_number, ''),
		          COALESCE(source_environment, ''), COALESCE(safety_level, ''), is_model_organism,
		          COALESCE(summary, ''), status, data_quality_score, created_at, updated_at, published_at
	`

	item, err := scanSpecies(repo.pool.QueryRow(ctx, query,
		input.Slug,
		input.LatinName,
		input.ChineseName,
		input.StrainNumber,
		input.SourceEnvironment,
		input.SafetyLevel,
		input.IsModelOrganism,
		input.Summary,
	))
	if isUniqueViolation(err) {
		return Species{}, ErrDuplicateSlug
	}
	return item, err
}

func (repo *PostgresRepository) Update(ctx context.Context, idOrSlug string, input UpdateInput) (Species, error) {
	query := `
		UPDATE species
		SET slug = $2,
		    latin_name = $3,
		    chinese_name = NULLIF($4, ''),
		    strain_number = NULLIF($5, ''),
		    source_environment = NULLIF($6, ''),
		    safety_level = NULLIF($7, ''),
		    is_model_organism = $8,
		    summary = NULLIF($9, ''),
		    status = CASE WHEN status = 'pending_review' THEN status ELSE 'draft' END,
		    published_at = NULL
		WHERE id::text = $1 OR slug = $1
		RETURNING id::text, slug, latin_name, COALESCE(chinese_name, ''), COALESCE(strain_number, ''),
		          COALESCE(source_environment, ''), COALESCE(safety_level, ''), is_model_organism,
		          COALESCE(summary, ''), status, data_quality_score, created_at, updated_at, published_at
	`

	item, err := scanSpecies(repo.pool.QueryRow(ctx, query,
		idOrSlug,
		input.Slug,
		input.LatinName,
		input.ChineseName,
		input.StrainNumber,
		input.SourceEnvironment,
		input.SafetyLevel,
		input.IsModelOrganism,
		input.Summary,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return Species{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return Species{}, ErrDuplicateSlug
	}
	return item, err
}

func (repo *PostgresRepository) Archive(ctx context.Context, idOrSlug string) error {
	command, err := repo.pool.Exec(ctx, `UPDATE species SET status = 'archived', published_at = NULL WHERE id::text = $1 OR slug = $1`, idOrSlug)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (repo *PostgresRepository) Delete(ctx context.Context, idOrSlug string) error {
	command, err := repo.pool.Exec(ctx, `DELETE FROM species WHERE id::text = $1 OR slug = $1`, idOrSlug)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSpecies(row scanner) (Species, error) {
	var item Species
	err := row.Scan(
		&item.ID,
		&item.Slug,
		&item.LatinName,
		&item.ChineseName,
		&item.StrainNumber,
		&item.SourceEnvironment,
		&item.SafetyLevel,
		&item.IsModelOrganism,
		&item.Summary,
		&item.Status,
		&item.DataQualityScore,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.PublishedAt,
	)
	return item, err
}

func normalizeStatus(status Status) Status {
	switch status {
	case StatusPendingReview, StatusPublished, StatusArchived:
		return status
	default:
		return StatusDraft
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
