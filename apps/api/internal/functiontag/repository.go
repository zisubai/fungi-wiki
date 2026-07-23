package functiontag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("function tag not found")
var ErrDuplicateCode = errors.New("function tag code already exists")

type Repository interface {
	List(ctx context.Context, params ListParams) ([]FunctionTag, error)
	Get(ctx context.Context, idOrCode string) (FunctionTag, error)
	Create(ctx context.Context, input CreateInput) (FunctionTag, error)
	Update(ctx context.Context, idOrCode string, input UpdateInput) (FunctionTag, error)
	Delete(ctx context.Context, idOrCode string) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (repo *PostgresRepository) List(ctx context.Context, params ListParams) ([]FunctionTag, error) {
	limit := params.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	where := []string{"1 = 1"}
	args := make([]any, 0)
	if params.Query != "" {
		args = append(args, "%"+strings.TrimSpace(params.Query)+"%")
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d OR description ILIKE $%d)", len(args), len(args), len(args)))
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(`
		SELECT ft.id::text, COALESCE(ft.parent_id::text, ''), ft.name, ft.code, COALESCE(ft.description, ''), ft.sort_order, ft.created_at, ft.updated_at,
		       (SELECT COUNT(*) FROM species_functions sf JOIN species s ON s.id=sf.species_id WHERE sf.function_tag_id=ft.id AND s.status='published')
		FROM function_tags ft
		WHERE %s
		ORDER BY ft.sort_order ASC, ft.name ASC
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), len(args)-1, len(args))

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]FunctionTag, 0)
	for rows.Next() {
		item, err := scanFunctionTag(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *PostgresRepository) Get(ctx context.Context, idOrCode string) (FunctionTag, error) {
	query := `
		SELECT ft.id::text, COALESCE(ft.parent_id::text, ''), ft.name, ft.code, COALESCE(ft.description, ''), ft.sort_order, ft.created_at, ft.updated_at,
		       (SELECT COUNT(*) FROM species_functions sf JOIN species s ON s.id=sf.species_id WHERE sf.function_tag_id=ft.id AND s.status='published')
		FROM function_tags ft
		WHERE ft.id::text = $1 OR ft.code = $1
		LIMIT 1
	`

	item, err := scanFunctionTag(repo.pool.QueryRow(ctx, query, idOrCode))
	if errors.Is(err, pgx.ErrNoRows) {
		return FunctionTag{}, ErrNotFound
	}
	return item, err
}

func (repo *PostgresRepository) Create(ctx context.Context, input CreateInput) (FunctionTag, error) {
	query := `
		INSERT INTO function_tags (parent_id, name, code, description, sort_order)
		VALUES (NULLIF($1, '')::uuid, $2, $3, NULLIF($4, ''), $5)
		RETURNING id::text, COALESCE(parent_id::text, ''), name, code, COALESCE(description, ''), sort_order, created_at, updated_at, 0
	`

	item, err := scanFunctionTag(repo.pool.QueryRow(ctx, query, input.ParentID, input.Name, input.Code, input.Description, input.SortOrder))
	if isUniqueViolation(err) {
		return FunctionTag{}, ErrDuplicateCode
	}
	return item, err
}

func (repo *PostgresRepository) Update(ctx context.Context, idOrCode string, input UpdateInput) (FunctionTag, error) {
	query := `
		UPDATE function_tags
		SET parent_id = NULLIF($2, '')::uuid,
		    name = $3,
		    code = $4,
		    description = NULLIF($5, ''),
		    sort_order = $6
		WHERE id::text = $1 OR code = $1
		RETURNING id::text, COALESCE(parent_id::text, ''), name, code, COALESCE(description, ''), sort_order, created_at, updated_at, 0
	`

	item, err := scanFunctionTag(repo.pool.QueryRow(ctx, query, idOrCode, input.ParentID, input.Name, input.Code, input.Description, input.SortOrder))
	if errors.Is(err, pgx.ErrNoRows) {
		return FunctionTag{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return FunctionTag{}, ErrDuplicateCode
	}
	return item, err
}

func (repo *PostgresRepository) Delete(ctx context.Context, idOrCode string) error {
	command, err := repo.pool.Exec(ctx, `DELETE FROM function_tags WHERE id::text = $1 OR code = $1`, idOrCode)
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

func scanFunctionTag(row scanner) (FunctionTag, error) {
	var item FunctionTag
	err := row.Scan(
		&item.ID,
		&item.ParentID,
		&item.Name,
		&item.Code,
		&item.Description,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.PublishedSpeciesCount,
	)
	return item, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
