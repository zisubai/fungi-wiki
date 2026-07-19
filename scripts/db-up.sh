#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONTAINER_NAME="fungi-wiki-postgres"
DB_PORT="${DB_PORT:-55432}"

if ! docker info >/dev/null 2>&1; then
  echo "Docker daemon is not running. Please start Docker Desktop and retry." >&2
  exit 1
fi

if docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER_NAME"; then
  docker start "$CONTAINER_NAME" >/dev/null
else
  docker run -d \
    --name "$CONTAINER_NAME" \
    -e POSTGRES_DB=fungi_wiki \
    -e POSTGRES_USER=fungi \
    -e POSTGRES_PASSWORD=fungi \
    -p "${DB_PORT}:5432" \
    -v "${ROOT_DIR}/apps/api/migrations:/docker-entrypoint-initdb.d:ro" \
    postgres:17-alpine >/dev/null
fi

echo "PostgreSQL is starting on localhost:${DB_PORT}"
echo "DATABASE_URL=postgres://fungi:fungi@localhost:${DB_PORT}/fungi_wiki?sslmode=disable"
