#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MODULE_ROOT="$SKILL_ROOT/drawcli"
OUTPUT_DIR="${1:-$SKILL_ROOT/../bins}"

mkdir -p "$OUTPUT_DIR"
cd "$MODULE_ROOT"

build_one() {
  local goos="$1"
  local goarch="$2"
  local suffix="$3"
  local output="$OUTPUT_DIR/drawcli-${goos}-${goarch}${suffix}"
  echo "building $output"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o "$output" .
}

# -s removes the symbol table; -w removes DWARF debug data.
build_one linux amd64 ""
build_one linux arm64 ""
build_one windows amd64 ".exe"
build_one darwin amd64 ""
build_one darwin arm64 ""

echo "built drawcli binaries in $OUTPUT_DIR"
