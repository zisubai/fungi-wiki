#!/usr/bin/env bash
set -euo pipefail

container_name="${POSTGRES_CONTAINER:-fungi-wiki-postgres}"
database_name="${POSTGRES_DB:-fungi_wiki}"
database_user="${POSTGRES_USER:-fungi}"
backup_directory="${BACKUP_DIR:-backups}"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_path="${backup_directory}/${database_name}-${timestamp}.dump"

if ! docker inspect "$container_name" >/dev/null 2>&1; then
  echo "PostgreSQL container not found: $container_name" >&2
  exit 1
fi

if [ "$(docker inspect -f '{{.State.Running}}' "$container_name")" != "true" ]; then
  echo "PostgreSQL container is not running: $container_name" >&2
  exit 1
fi

mkdir -p "$backup_directory"
umask 077

docker exec "$container_name" pg_dump \
  --username "$database_user" \
  --dbname "$database_name" \
  --format custom \
  --no-owner \
  --no-privileges > "$backup_path"

if [ ! -s "$backup_path" ]; then
  echo "Backup is empty: $backup_path" >&2
  exit 1
fi

docker exec -i "$container_name" pg_restore --list < "$backup_path" >/dev/null

echo "$backup_path"
