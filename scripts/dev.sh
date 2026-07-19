#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if ! command -v npm >/dev/null 2>&1; then
  echo "npm is required." >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "go is required." >&2
  exit 1
fi

cleanup() {
  if [[ -n "${API_PID:-}" ]]; then kill "$API_PID" 2>/dev/null || true; fi
  if [[ -n "${WEB_PID:-}" ]]; then kill "$WEB_PID" 2>/dev/null || true; fi
  if [[ -n "${ADMIN_PID:-}" ]]; then kill "$ADMIN_PID" 2>/dev/null || true; fi
}
trap cleanup EXIT INT TERM

if docker info >/dev/null 2>&1; then
  ./scripts/db-up.sh
else
  echo "Docker daemon is not ready; skipping database startup."
  echo "If PostgreSQL is not already running, start Docker Desktop and run ./scripts/db-up.sh."
fi

export DATABASE_URL="${DATABASE_URL:-postgres://fungi:fungi@localhost:55432/fungi_wiki?sslmode=disable}"
export VITE_API_BASE_URL="${VITE_API_BASE_URL:-http://localhost:8080}"

npm run dev:api & API_PID=$!
npm run dev:web & WEB_PID=$!
npm run dev:admin & ADMIN_PID=$!

echo ""
echo "Development services started:"
echo "- API:   http://localhost:8080"
echo "- Web:   http://localhost:5173"
echo "- Admin: http://localhost:5174"
echo ""
echo "Press Ctrl+C to stop all services."

wait
