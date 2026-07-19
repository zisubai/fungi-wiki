package species

import (
	"context"
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
	List(ctx context.Context, params ListParams) ([]Species, error)
	Get(ctx context.Context, idOrSlug string) (Species, error)
	Create(ctx context.Context, input CreateInput) (Species, error)
	Update(ctx context.Context, idOrSlug string, input UpdateInput) (Species, error)
	Archive(ctx context.Context, idOrSlug string) error
	Delete(ctx context.Context, idOrSlug string) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repo *PostgresRepository) List(ctx context.Context, params ListParams) ([]Species, error) {
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

	if params.Query != "" {
		args = append(args, "%"+strings.TrimSpace(params.Query)+"%")
		where = append(where, fmt.Sprintf("(slug ILIKE $%d OR latin_name ILIKE $%d OR chinese_name ILIKE $%d OR summary ILIKE $%d)", len(args), len(args), len(args), len(args)))
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

	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT id::text, slug, latin_name, COALESCE(chinese_name, ''), COALESCE(strain_number, ''),
		       COALESCE(source_environment, ''), COALESCE(safety_level, ''), is_model_organism,
		       COALESCE(summary, ''), status, data_quality_score, created_at, updated_at, published_at
		FROM species
		WHERE %s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), len(args)-1, len(args))

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Species, 0)
	for rows.Next() {
		item, err := scanSpecies(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
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
