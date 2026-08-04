#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ ! -f .env.processor ]]; then
  cp .env.processor.example .env.processor
  echo "created .env from .env.example"
fi

set -a
# shellcheck disable=SC1091
source .env.processor
set +a

task docker:up
task deps
./scripts/migrate.sh
task build
task run