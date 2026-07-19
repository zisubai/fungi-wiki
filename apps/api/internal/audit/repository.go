package audit

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("audit record not found")
var ErrInvalidState = errors.New("invalid audit state")
var ErrSpeciesNotFound = errors.New("species not found")

type Repository interface {
	List(context.Context, string) ([]Record, error)
	Submit(context.Context, string) (Record, error)
	Review(context.Context, string, bool, string) (Record, error)
}
type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(p *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{p} }

const selectSQL = `SELECT a.id::text,a.entity_type,a.entity_id::text,COALESCE(s.latin_name,''),a.action,a.status,COALESCE(a.comment,''),a.submitted_at,a.reviewed_at FROM audit_records a LEFT JOIN species s ON a.entity_type='species' AND s.id=a.entity_id`

func scan(row interface{ Scan(...any) error }) (Record, error) {
	var x Record
	e := row.Scan(&x.ID, &x.EntityType, &x.EntityID, &x.EntityName, &x.Action, &x.Status, &x.Comment, &x.SubmittedAt, &x.ReviewedAt)
	return x, e
}
func (r *PostgresRepository) List(ctx context.Context, status string) ([]Record, error) {
	args := []any{}
	q := selectSQL
	if status != "" {
		q += ` WHERE a.status=$1`
		args = append(args, status)
	}
	q += ` ORDER BY a.submitted_at DESC`
	rows, e := r.pool.Query(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	xs := []Record{}
	for rows.Next() {
		x, e := scan(rows)
		if e != nil {
			return nil, e
		}
		xs = append(xs, x)
	}
	return xs, rows.Err()
}
func (r *PostgresRepository) Submit(ctx context.Context, v string) (Record, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Record{}, e
	}
	defer tx.Rollback(ctx)
	var sid string
	e = tx.QueryRow(ctx, `SELECT id::text FROM species WHERE (id::text=$1 OR slug=$1) AND status IN('draft','archived') FOR UPDATE`, v).Scan(&sid)
	if errors.Is(e, pgx.ErrNoRows) {
		return Record{}, ErrSpeciesNotFound
	}
	if e != nil {
		return Record{}, e
	}
	var id string
	e = tx.QueryRow(ctx, `INSERT INTO audit_records(entity_type,entity_id,action,status)VALUES('species',$1::uuid,'publish','pending')RETURNING id::text`, sid).Scan(&id)
	if e != nil {
		return Record{}, e
	}
	_, e = tx.Exec(ctx, `UPDATE species SET status='pending_review',published_at=NULL WHERE id=$1::uuid`, sid)
	if e != nil {
		return Record{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return Record{}, e
	}
	return scan(r.pool.QueryRow(ctx, selectSQL+` WHERE a.id::text=$1`, id))
}
func (r *PostgresRepository) Review(ctx context.Context, id string, approve bool, comment string) (Record, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Record{}, e
	}
	defer tx.Rollback(ctx)
	var sid, status string
	e = tx.QueryRow(ctx, `SELECT entity_id::text,status FROM audit_records WHERE id::text=$1 AND entity_type='species' FOR UPDATE`, id).Scan(&sid, &status)
	if errors.Is(e, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	if e != nil {
		return Record{}, e
	}
	if status != "pending" {
		return Record{}, ErrInvalidState
	}
	nextAudit, nextSpecies := "rejected", "draft"
	if approve {
		nextAudit, nextSpecies = "approved", "published"
	}
	_, e = tx.Exec(ctx, `UPDATE audit_records SET status=$2,comment=NULLIF($3,''),reviewed_at=NOW() WHERE id::text=$1`, id, nextAudit, comment)
	if e != nil {
		return Record{}, e
	}
	_, e = tx.Exec(ctx, `UPDATE species SET status=$2,published_at=CASE WHEN $2='published' THEN NOW() ELSE NULL END WHERE id=$1::uuid`, sid, nextSpecies)
	if e != nil {
		return Record{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return Record{}, e
	}
	return scan(r.pool.QueryRow(ctx, selectSQL+` WHERE a.id::text=$1`, id))
}
