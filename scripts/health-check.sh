#!/usr/bin/env bash
set -euo pipefail

api_url="${API_URL:-http://localhost:8080}"
web_url="${WEB_URL:-http://localhost:5173}"
admin_url="${ADMIN_URL:-http://localhost:5174}"
postgres_container="${POSTGRES_CONTAINER:-fungi-wiki-postgres}"
postgres_database="${POSTGRES_DB:-fungi_wiki}"
postgres_user="${POSTGRES_USER:-fungi}"
timeout_seconds="${HEALTH_CHECK_TIMEOUT:-10}"

check_http() {
  local name="$1"
  local url="$2"
  curl \
    --fail \
    --silent \
    --show-error \
    --connect-timeout "$timeout_seconds" \
    --max-time "$timeout_seconds" \
    "$url" >/dev/null
  printf 'ok  %s  %s\n' "$name" "$url"
}

if [ "$(docker inspect -f '{{.State.Running}}' "$postgres_container" 2>/dev/null)" != "true" ]; then
  echo "fail  postgres container is not running: $postgres_container" >&2
  exit 1
fi

docker exec "$postgres_container" pg_isready \
  --username "$postgres_user" \
  --dbname "$postgres_database" >/dev/null
printf 'ok  postgres  container=%s database=%s\n' "$postgres_container" "$postgres_database"

check_http "api-live" "${api_url}/healthz"
check_http "api-ready" "${api_url}/readyz"
check_http "public-species" "${api_url}/api/species?limit=1"
check_http "web" "${web_url}/healthz"
check_http "admin" "${admin_url}/healthz"

echo "All health checks passed."
