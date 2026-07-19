package migrations

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var files embed.FS

const migrationLockID int64 = 672019031

type Applied struct {
	Name      string
	Checksum  string
	AppliedAt time.Time
}

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer connection.Release()
	if _, err = connection.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLockID); err != nil {
		return err
	}
	defer connection.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, migrationLockID)
	if _, err = connection.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(name TEXT PRIMARY KEY,checksum TEXT NOT NULL,applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return err
	}
	names, err := Names()
	if err != nil {
		return err
	}
	if err = baselineExisting(ctx, connection, names); err != nil {
		return err
	}
	for _, name := range names {
		content, readErr := files.ReadFile(name)
		if readErr != nil {
			return readErr
		}
		checksum := checksumOf(content)
		var stored string
		err = connection.QueryRow(ctx, `SELECT checksum FROM schema_migrations WHERE name=$1`, name).Scan(&stored)
		if err == nil {
			if stored != checksum {
				return fmt.Errorf("migration %s was modified after being applied", name)
			}
			continue
		}
		tx, beginErr := connection.Begin(ctx)
		if beginErr != nil {
			return beginErr
		}
		if _, execErr := tx.Exec(ctx, string(content)); execErr != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", name, execErr)
		}
		if _, execErr := tx.Exec(ctx, `INSERT INTO schema_migrations(name,checksum)VALUES($1,$2)`, name, checksum); execErr != nil {
			tx.Rollback(ctx)
			return execErr
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return commitErr
		}
	}
	return nil
}

func Names() ([]string, error) {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// baselineExisting recognizes databases created before the migration runner existed.
func baselineExisting(ctx context.Context, connection *pgxpool.Conn, names []string) error {
	detectors := map[string]string{
		"001_init_schema.sql":     `SELECT to_regclass('public.species') IS NOT NULL`,
		"002_search_indexes.sql":  `SELECT to_regclass('public.idx_species_source_environment_trgm') IS NOT NULL`,
		"003_import_batches.sql":  `SELECT to_regclass('public.import_batches') IS NOT NULL`,
		"004_users_and_roles.sql": `SELECT to_regclass('public.users') IS NOT NULL`,
	}
	for _, name := range names {
		detector, ok := detectors[name]
		if !ok {
			continue
		}
		var applied bool
		if err := connection.QueryRow(ctx, detector).Scan(&applied); err != nil {
			return err
		}
		if !applied {
			continue
		}
		content, err := files.ReadFile(name)
		if err != nil {
			return err
		}
		if _, err = connection.Exec(ctx, `INSERT INTO schema_migrations(name,checksum)VALUES($1,$2)ON CONFLICT(name)DO NOTHING`, name, checksumOf(content)); err != nil {
			return err
		}
	}
	return nil
}

func checksumOf(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
