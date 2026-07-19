package speciesfunction

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSpeciesNotFound = errors.New("species not found")

type Repository interface {
	List(ctx context.Context, speciesIDOrSlug string) ([]SpeciesFunction, error)
	Replace(ctx context.Context, speciesIDOrSlug string, items []ReplaceItem) ([]SpeciesFunction, error)
}

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repo *PostgresRepository) List(ctx context.Context, speciesIDOrSlug string) ([]SpeciesFunction, error) {
	var speciesID string
	err := repo.pool.QueryRow(ctx, `SELECT id::text FROM species WHERE id::text = $1 OR slug = $1`, speciesIDOrSlug).Scan(&speciesID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSpeciesNotFound
	}
	if err != nil {
		return nil, err
	}
	return listBySpeciesID(ctx, repo.pool, speciesID)
}

func (repo *PostgresRepository) Replace(ctx context.Context, speciesIDOrSlug string, items []ReplaceItem) ([]SpeciesFunction, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var speciesID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM species WHERE id::text = $1 OR slug = $1 FOR UPDATE`, speciesIDOrSlug).Scan(&speciesID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSpeciesNotFound
	}
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM species_functions WHERE species_id = $1::uuid`, speciesID); err != nil {
		return nil, err
	}
	for _, item := range items {
		_, err = tx.Exec(ctx, `
			INSERT INTO species_functions
			(species_id, function_tag_id, description, function_strength, verification_method, applicable_environment, confidence_score)
			VALUES ($1::uuid, $2::uuid, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7)
		`, speciesID, item.FunctionTagID, item.Description, item.FunctionStrength, item.VerificationMethod, item.ApplicableEnvironment, item.ConfidenceScore)
		if err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return repo.List(ctx, speciesID)
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func listBySpeciesID(ctx context.Context, db queryer, speciesID string) ([]SpeciesFunction, error) {
	rows, err := db.Query(ctx, `
		SELECT sf.id::text, sf.species_id::text, sf.function_tag_id::text, ft.name, ft.code,
		       COALESCE(sf.description, ''), COALESCE(sf.function_strength, ''),
		       COALESCE(sf.verification_method, ''), COALESCE(sf.applicable_environment, ''),
		       sf.confidence_score, sf.created_at, sf.updated_at
		FROM species_functions sf
		JOIN function_tags ft ON ft.id = sf.function_tag_id
		WHERE sf.species_id = $1::uuid
		ORDER BY ft.sort_order, ft.name
	`, speciesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SpeciesFunction, 0)
	for rows.Next() {
		var item SpeciesFunction
		if err := rows.Scan(&item.ID, &item.SpeciesID, &item.FunctionTagID, &item.FunctionTagName,
			&item.FunctionTagCode, &item.Description, &item.FunctionStrength, &item.VerificationMethod,
			&item.ApplicableEnvironment, &item.ConfidenceScore, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
