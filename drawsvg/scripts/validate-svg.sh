#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <svg-file>" >&2
  exit 1
fi

exec "$SCRIPT_DIR/drawcli" check "$1"
