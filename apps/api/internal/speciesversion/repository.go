package speciesversion

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSpeciesNotFound = errors.New("species not found")

type Repository interface {
	List(context.Context, string, int) ([]Version, error)
}
type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) List(ctx context.Context, value string, limit int) ([]Version, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var speciesID string
	if err := r.pool.QueryRow(ctx, `SELECT id::text FROM species WHERE id::text=$1 OR slug=$1`, value).Scan(&speciesID); errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSpeciesNotFound
	} else if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id::text,species_id::text,version_number,change_type,source_table,snapshot,created_at FROM species_versions WHERE species_id=$1::uuid ORDER BY version_number DESC LIMIT $2`, speciesID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Version{}
	for rows.Next() {
		var item Version
		if err := rows.Scan(&item.ID, &item.SpeciesID, &item.VersionNumber, &item.ChangeType, &item.SourceTable, &item.Snapshot, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
