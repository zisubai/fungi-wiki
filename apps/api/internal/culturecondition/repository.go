package culturecondition

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSpeciesNotFound = errors.New("species not found")

type Repository interface {
	List(context.Context, string) ([]Condition, error)
	Replace(context.Context, string, []Input) ([]Condition, error)
}
type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func resolveSpecies(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, value string) (string, error) {
	var id string
	err := q.QueryRow(ctx, `SELECT id::text FROM species WHERE id::text=$1 OR slug=$1`, value).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSpeciesNotFound
	}
	return id, err
}
func (r *PostgresRepository) List(ctx context.Context, value string) ([]Condition, error) {
	id, err := resolveSpecies(ctx, r.pool, value)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id::text,species_id::text,COALESCE(medium_name,''),temperature_min,temperature_max,ph_min,ph_max,salinity_min,salinity_max,COALESCE(oxygen_requirement,''),COALESCE(culture_time,''),COALESCE(notes,''),created_at,updated_at FROM culture_conditions WHERE species_id=$1::uuid ORDER BY created_at`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Condition{}
	for rows.Next() {
		var x Condition
		if err = rows.Scan(&x.ID, &x.SpeciesID, &x.MediumName, &x.TemperatureMin, &x.TemperatureMax, &x.PHMin, &x.PHMax, &x.SalinityMin, &x.SalinityMax, &x.OxygenRequirement, &x.CultureTime, &x.Notes, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *PostgresRepository) Replace(ctx context.Context, value string, inputs []Input) ([]Condition, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	id, err := resolveSpecies(ctx, tx, value)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM culture_conditions WHERE species_id=$1::uuid`, id); err != nil {
		return nil, err
	}
	for _, x := range inputs {
		_, err = tx.Exec(ctx, `INSERT INTO culture_conditions(species_id,medium_name,temperature_min,temperature_max,ph_min,ph_max,salinity_min,salinity_max,oxygen_requirement,culture_time,notes) VALUES($1::uuid,NULLIF($2,''),$3,$4,$5,$6,$7,$8,NULLIF($9,''),NULLIF($10,''),NULLIF($11,''))`, id, x.MediumName, x.TemperatureMin, x.TemperatureMax, x.PHMin, x.PHMax, x.SalinityMin, x.SalinityMax, x.OxygenRequirement, x.CultureTime, x.Notes)
		if err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.List(ctx, id)
}
