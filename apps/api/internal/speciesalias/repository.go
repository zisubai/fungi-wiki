package speciesalias

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSpeciesNotFound = errors.New("species not found")

type Repository interface {
	List(context.Context, string) ([]Alias, error)
	Replace(context.Context, string, []Input) ([]Alias, error)
}
type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(p *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{p} }
func resolve(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, value string) (string, error) {
	var id string
	err := q.QueryRow(ctx, `SELECT id::text FROM species WHERE id::text=$1 OR slug=$1`, value).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSpeciesNotFound
	}
	return id, err
}
func (r *PostgresRepository) List(ctx context.Context, value string) ([]Alias, error) {
	id, err := resolve(ctx, r.pool, value)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id::text,species_id::text,alias_name,alias_type,COALESCE(source,''),created_at FROM species_aliases WHERE species_id=$1::uuid ORDER BY alias_name`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Alias{}
	for rows.Next() {
		var x Alias
		if err = rows.Scan(&x.ID, &x.SpeciesID, &x.Name, &x.Type, &x.Source, &x.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *PostgresRepository) Replace(ctx context.Context, value string, inputs []Input) ([]Alias, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	id, err := resolve(ctx, tx, value)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM species_aliases WHERE species_id=$1::uuid`, id); err != nil {
		return nil, err
	}
	for _, x := range inputs {
		kind := x.Type
		if kind == "" {
			kind = "synonym"
		}
		if _, err = tx.Exec(ctx, `INSERT INTO species_aliases(species_id,alias_name,alias_type,source)VALUES($1::uuid,$2,$3,NULLIF($4,''))`, id, x.Name, kind, x.Source); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.List(ctx, id)
}
