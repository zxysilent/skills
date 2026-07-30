#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <svg-file>" >&2
  exit 1
fi

if [ -x "$SCRIPT_DIR/drawcli" ]; then
  exec "$SCRIPT_DIR/drawcli" check "$1"
fi

echo "drawcli is not installed; run scripts/install-drawcli.sh first" >&2
exit 1
