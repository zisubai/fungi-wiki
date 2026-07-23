#!/usr/bin/env bash
set -euo pipefail

container_name="${POSTGRES_CONTAINER:-fungi-wiki-postgres}"
source_database="${POSTGRES_DB:-fungi_wiki}"
database_user="${POSTGRES_USER:-fungi}"
backup_path="${1:-}"

if [ -z "$backup_path" ]; then
  shopt -s nullglob
  backup_candidates=(backups/*.dump)
  shopt -u nullglob
  if [ "${#backup_candidates[@]}" -eq 0 ]; then
    echo "No backup found under backups/. Run ./scripts/db-backup.sh first." >&2
    exit 1
  fi
  backup_path="$(ls -1t -- "${backup_candidates[@]}" | head -n 1)"
fi

if [ ! -r "$backup_path" ]; then
  echo "Backup is not readable: $backup_path" >&2
  exit 1
fi

if [ "$(docker inspect -f '{{.State.Running}}' "$container_name" 2>/dev/null)" != "true" ]; then
  echo "PostgreSQL container is not running: $container_name" >&2
  exit 1
fi

drill_database="${source_database}_restore_drill_$(date -u +%Y%m%dT%H%M%SZ)_$$"

cleanup() {
  docker exec "$container_name" dropdb \
    --username "$database_user" \
    --if-exists "$drill_database" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

docker exec -i "$container_name" pg_restore --list < "$backup_path" >/dev/null
docker exec "$container_name" createdb --username "$database_user" "$drill_database"
docker exec -i "$container_name" pg_restore \
  --username "$database_user" \
  --dbname "$drill_database" \
  --no-owner \
  --no-privileges < "$backup_path"

source_tables="$(docker exec "$container_name" psql \
  --username "$database_user" \
  --dbname "$source_database" \
  --tuples-only --no-align \
  --command "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename")"
restored_tables="$(docker exec "$container_name" psql \
  --username "$database_user" \
  --dbname "$drill_database" \
  --tuples-only --no-align \
  --command "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename")"

if [ "$source_tables" != "$restored_tables" ]; then
  echo "Public table lists differ between source and restored databases." >&2
  exit 1
fi

table_count=0
while IFS= read -r table_name; do
  [ -n "$table_name" ] || continue
  quoted_table="${table_name//\"/\"\"}"
  source_count="$(docker exec "$container_name" psql --username "$database_user" --dbname "$source_database" --tuples-only --no-align --command "SELECT count(*) FROM \"$quoted_table\"")"
  restored_count="$(docker exec "$container_name" psql --username "$database_user" --dbname "$drill_database" --tuples-only --no-align --command "SELECT count(*) FROM \"$quoted_table\"")"
  if [ "$source_count" != "$restored_count" ]; then
    echo "Row count mismatch for $table_name: source=$source_count restored=$restored_count" >&2
    exit 1
  fi
  table_count=$((table_count + 1))
done <<< "$source_tables"

echo "Restore drill passed: $backup_path ($table_count public tables verified)"
