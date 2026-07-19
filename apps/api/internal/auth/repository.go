package auth

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrDisabled = errors.New("user disabled")

type Repository interface {
	Authenticate(context.Context, string, string) (User, error)
	Get(context.Context, string) (User, error)
	List(context.Context) ([]User, error)
	Create(context.Context, CreateUserInput) (User, error)
	BootstrapAdmin(context.Context, string, string) error
}
type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(p *pgxpool.Pool) *PostgresRepository { return &PostgresRepository{p} }

const userColumns = `id::text,email,password_hash,display_name,role,status,last_login_at`

func scan(row interface{ Scan(...any) error }) (User, error) {
	var u User
	e := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Status, &u.LastLoginAt)
	return u, e
}
func (r *PostgresRepository) Authenticate(ctx context.Context, email, password string) (User, error) {
	u, e := scan(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email=$1`, strings.ToLower(strings.TrimSpace(email))))
	if errors.Is(e, pgx.ErrNoRows) {
		return User{}, ErrInvalidCredentials
	}
	if e != nil {
		return User{}, e
	}
	if u.Status != "active" {
		return User{}, ErrDisabled
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return User{}, ErrInvalidCredentials
	}
	_, e = r.pool.Exec(ctx, `UPDATE users SET last_login_at=NOW() WHERE id=$1::uuid`, u.ID)
	return u, e
}
func (r *PostgresRepository) Get(ctx context.Context, id string) (User, error) {
	u, e := scan(r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id=$1::uuid`, id))
	if errors.Is(e, pgx.ErrNoRows) {
		return User{}, ErrInvalidCredentials
	}
	return u, e
}
func (r *PostgresRepository) List(ctx context.Context) ([]User, error) {
	rows, e := r.pool.Query(ctx, `SELECT `+userColumns+` FROM users ORDER BY created_at`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		user, e := scan(rows)
		if e != nil {
			return nil, e
		}
		items = append(items, user)
	}
	return items, rows.Err()
}
func (r *PostgresRepository) Create(ctx context.Context, input CreateUserInput) (User, error) {
	hash, e := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if e != nil {
		return User{}, e
	}
	return scan(r.pool.QueryRow(ctx, `INSERT INTO users(email,password_hash,display_name,role,status)VALUES($1,$2,$3,$4,'active')RETURNING `+userColumns, strings.ToLower(strings.TrimSpace(input.Email)), string(hash), input.DisplayName, input.Role))
}
func (r *PostgresRepository) BootstrapAdmin(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		return nil
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		return e
	}
	_, e = r.pool.Exec(ctx, `INSERT INTO users(email,password_hash,display_name,role,status)VALUES($1,$2,'系统管理员','admin','active')ON CONFLICT(email)DO UPDATE SET password_hash=EXCLUDED.password_hash`, strings.ToLower(strings.TrimSpace(email)), string(hash))
	return e
}
