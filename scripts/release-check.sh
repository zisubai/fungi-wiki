#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[1/3] Running tests, static checks, and production builds"
npm run verify

echo "[2/3] Validating Docker Compose delivery configuration"
if docker compose version >/dev/null 2>&1; then
  docker compose config --quiet
elif command -v docker-compose >/dev/null 2>&1; then
  docker-compose config --quiet
else
  echo "Docker Compose is not installed; configuration validation skipped."
fi

echo "[3/3] Checking running services when available"
if curl --fail --silent --max-time 2 http://localhost:8080/healthz >/dev/null 2>&1; then
  ./scripts/smoke-api.sh
  curl --fail --silent --max-time 2 http://localhost:5173/ >/dev/null
  curl --fail --silent --max-time 2 http://localhost:5174/ >/dev/null
  echo "Runtime acceptance checks passed."
else
  echo "API is not running; runtime checks skipped. Start with ./scripts/dev.sh or docker compose up -d --build."
fi

echo "Release checks passed."
