package applicationcase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("application case not found")
var ErrSpeciesNotFound = errors.New("species not found")

type Repository interface {
	List(context.Context, string) ([]Case, error)
	Create(context.Context, string, Input) (Case, error)
	Update(context.Context, string, string, Input) (Case, error)
	Delete(context.Context, string, string) error
}

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) speciesID(ctx context.Context, value string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id::text FROM species WHERE id::text=$1 OR slug=$1`, value).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSpeciesNotFound
	}
	return id, err
}

const fields = `id::text,species_id::text,industry,scenario,COALESCE(problem,''),COALESCE(solution,''),COALESCE(result_summary,''),COALESCE(maturity_level,''),COALESCE(source,''),created_at,updated_at`

func scan(row interface{ Scan(...any) error }) (Case, error) {
	var item Case
	err := row.Scan(&item.ID, &item.SpeciesID, &item.Industry, &item.Scenario, &item.Problem, &item.Solution, &item.ResultSummary, &item.MaturityLevel, &item.Source, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *PostgresRepository) List(ctx context.Context, value string) ([]Case, error) {
	id, err := r.speciesID(ctx, value)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT `+fields+` FROM application_cases WHERE species_id=$1::uuid ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Case{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) Create(ctx context.Context, value string, input Input) (Case, error) {
	id, err := r.speciesID(ctx, value)
	if err != nil {
		return Case{}, err
	}
	return scan(r.pool.QueryRow(ctx, `INSERT INTO application_cases(species_id,industry,scenario,problem,solution,result_summary,maturity_level,source) VALUES($1::uuid,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,'')) RETURNING `+fields, id, input.Industry, input.Scenario, input.Problem, input.Solution, input.ResultSummary, input.MaturityLevel, input.Source))
}

func (r *PostgresRepository) Update(ctx context.Context, value, caseID string, input Input) (Case, error) {
	id, err := r.speciesID(ctx, value)
	if err != nil {
		return Case{}, err
	}
	item, err := scan(r.pool.QueryRow(ctx, `UPDATE application_cases SET industry=$3,scenario=$4,problem=NULLIF($5,''),solution=NULLIF($6,''),result_summary=NULLIF($7,''),maturity_level=NULLIF($8,''),source=NULLIF($9,'') WHERE id::text=$1 AND species_id=$2::uuid RETURNING `+fields, caseID, id, input.Industry, input.Scenario, input.Problem, input.Solution, input.ResultSummary, input.MaturityLevel, input.Source))
	if errors.Is(err, pgx.ErrNoRows) {
		return Case{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) Delete(ctx context.Context, value, caseID string) error {
	id, err := r.speciesID(ctx, value)
	if err != nil {
		return err
	}
	result, err := r.pool.Exec(ctx, `DELETE FROM application_cases WHERE id::text=$1 AND species_id=$2::uuid`, caseID, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
