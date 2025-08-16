#!/usr/bin/env bash
set -euo pipefail
if ! command -v air &> /dev/null; then
  echo "Please install air (https://github.com/air-verse/air) for hot reload."
  exit 1
fi
air
