#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="${DRAWCLI_BASE_URL:-https://raw.githubusercontent.com/zxysilent/skills/master/bins}"

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "unsupported operating system: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

url="$BASE_URL/drawcli-${os}-${arch}"
output="$SCRIPT_DIR/drawcli"
tmp="$output.tmp.$$"
trap 'rm -f "$tmp"' EXIT

if command -v curl >/dev/null 2>&1; then
  curl --fail --location --silent --show-error "$url" -o "$tmp"
elif command -v wget >/dev/null 2>&1; then
  wget --https-only --quiet "$url" -O "$tmp"
else
  echo "install curl or wget first" >&2
  exit 1
fi

chmod 755 "$tmp"
mv -f "$tmp" "$output"
echo "installed $url -> $output"
