#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"

headers_file="$(mktemp)"
trap 'rm -f "$headers_file"' EXIT
health_payload="$(curl --fail --silent --show-error --dump-header "$headers_file" "${API_BASE_URL}/healthz")"
if [[ "$health_payload" != *'"status":"ok"'* ]]; then
  echo "Unexpected health response: ${health_payload}" >&2
  exit 1
fi
if ! grep -qi '^X-Request-ID:' "$headers_file"; then
  echo "Health response is missing X-Request-ID" >&2
  exit 1
fi

ready_payload="$(curl --fail --silent --show-error "${API_BASE_URL}/readyz")"
if [[ "$ready_payload" != *'"status":"ready"'* ]]; then
  echo "Unexpected readiness response: ${ready_payload}" >&2
  exit 1
fi

curl --fail --silent --show-error "${API_BASE_URL}/api/species?limit=1" >/dev/null
echo "API smoke checks passed: ${API_BASE_URL}"
