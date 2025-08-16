#!/usr/bin/env bash
set -euo pipefail
if [ -f ".env" ]; then
  export $(grep -v '^#' .env | xargs)
fi
# Migration dijalankan otomatis saat start aplikasi.
echo "Migrations are executed on app start via internal/db/migrate.go"
