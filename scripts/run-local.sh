#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "created .env from .env.example"
fi

set -a
# shellcheck disable=SC1091
source .env
set +a

task docker:up
task deps
task build
task run