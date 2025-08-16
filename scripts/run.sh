#!/usr/bin/env bash
set -euo pipefail
if [ -f ".env" ]; then
  export $(grep -v '^#' .env | xargs)
fi
go run ./cmd/api
