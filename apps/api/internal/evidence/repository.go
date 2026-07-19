package evidence

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("evidence not found")
var ErrSpeciesNotFound = errors.New("species not found")

type Repository interface {
	List(context.Context, string) ([]Evidence, error)
	Create(context.Context, string, CreateInput) (Evidence, error)
	Delete(context.Context, string, string) error
}
type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(p *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{p} }
func resolve(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, v string) (string, error) {
	var id string
	e := q.QueryRow(ctx, `SELECT id::text FROM species WHERE id::text=$1 OR slug=$1`, v).Scan(&id)
	if errors.Is(e, pgx.ErrNoRows) {
		return "", ErrSpeciesNotFound
	}
	return id, e
}

const listSQL = `SELECT e.id::text,e.species_id::text,COALESCE(e.function_tag_id::text,''),COALESCE(ft.name,''),l.id::text,l.title,COALESCE(l.authors,''),COALESCE(l.journal,''),l.publication_year,COALESCE(l.doi,''),COALESCE(l.pmid,''),COALESCE(l.source_url,''),COALESCE(l.abstract,''),e.conclusion,e.evidence_level,e.evidence_score,e.created_at,e.updated_at FROM evidences e JOIN literatures l ON l.id=e.literature_id LEFT JOIN function_tags ft ON ft.id=e.function_tag_id WHERE e.species_id=$1::uuid ORDER BY e.created_at DESC`

func scan(row interface{ Scan(...any) error }) (Evidence, error) {
	var x Evidence
	e := row.Scan(&x.ID, &x.SpeciesID, &x.FunctionTagID, &x.FunctionTagName, &x.LiteratureID, &x.Title, &x.Authors, &x.Journal, &x.PublicationYear, &x.DOI, &x.PMID, &x.SourceURL, &x.Abstract, &x.Conclusion, &x.EvidenceLevel, &x.EvidenceScore, &x.CreatedAt, &x.UpdatedAt)
	return x, e
}
func (r *PostgresRepository) List(ctx context.Context, v string) ([]Evidence, error) {
	id, e := resolve(ctx, r.pool, v)
	if e != nil {
		return nil, e
	}
	rows, e := r.pool.Query(ctx, listSQL, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	xs := []Evidence{}
	for rows.Next() {
		x, e := scan(rows)
		if e != nil {
			return nil, e
		}
		xs = append(xs, x)
	}
	return xs, rows.Err()
}
func (r *PostgresRepository) Create(ctx context.Context, v string, in CreateInput) (Evidence, error) {
	tx, e := r.pool.Begin(ctx)
	if e != nil {
		return Evidence{}, e
	}
	defer tx.Rollback(ctx)
	sid, e := resolve(ctx, tx, v)
	if e != nil {
		return Evidence{}, e
	}
	level := in.EvidenceLevel
	if level == "" {
		level = "medium"
	}
	var lid, eid string
	e = tx.QueryRow(ctx, `INSERT INTO literatures(title,authors,journal,publication_year,doi,pmid,source_url,abstract)VALUES($1,NULLIF($2,''),NULLIF($3,''),$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,''))RETURNING id::text`, in.Title, in.Authors, in.Journal, in.PublicationYear, in.DOI, in.PMID, in.SourceURL, in.Abstract).Scan(&lid)
	if e != nil {
		return Evidence{}, e
	}
	e = tx.QueryRow(ctx, `INSERT INTO evidences(species_id,function_tag_id,literature_id,conclusion,evidence_level,evidence_score)VALUES($1::uuid,NULLIF($2,'')::uuid,$3::uuid,$4,$5,$6)RETURNING id::text`, sid, in.FunctionTagID, lid, in.Conclusion, level, in.EvidenceScore).Scan(&eid)
	if e != nil {
		return Evidence{}, e
	}
	if e = tx.Commit(ctx); e != nil {
		return Evidence{}, e
	}
	xs, e := r.List(ctx, sid)
	if e != nil {
		return Evidence{}, e
	}
	for _, x := range xs {
		if x.ID == eid {
			return x, nil
		}
	}
	return Evidence{}, ErrNotFound
}
func (r *PostgresRepository) Delete(ctx context.Context, v, id string) error {
	sid, e := resolve(ctx, r.pool, v)
	if e != nil {
		return e
	}
	cmd, e := r.pool.Exec(ctx, `DELETE FROM evidences WHERE id::text=$1 AND species_id=$2::uuid`, id, sid)
	if e != nil {
		return e
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
