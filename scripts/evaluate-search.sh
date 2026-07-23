#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../apps/api"
args=""
if [ "${REQUIRE_SEMANTIC:-0}" = "1" ]; then args="-require-semantic"; fi
go run ./cmd/search-eval -api "${API_URL:-http://localhost:8080}" -dataset "${SEARCH_EVAL_DATASET:-testdata/search-evaluation.json}" $args
