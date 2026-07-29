# Build DrawCLI

`drawcli` is the Go runtime for DrawSVG. Build release binaries from
`drawsvg/drawcli`; do not use the old Python generator as the runtime.

## Bundled Binaries

The skill includes:

- `bin/drawcli-linux-amd64`
- `bin/drawcli-linux-arm64`
- `bin/drawcli-windows-amd64.exe`
- `bin/drawcli-darwin-amd64`
- `bin/drawcli-darwin-arm64`

Use the binary matching the host platform. Windows and macOS binaries cannot be
smoke-tested on Linux without matching runners; their target formats can still
be checked with `file`.

## Rebuild Both Platforms

From any working directory:

```bash
SKILL_ROOT="${CLAUDE_SKILL_DIR:-/absolute/path/to/drawsvg}"
GOCACHE=/tmp/drawcli-cache "$SKILL_ROOT/scripts/build-drawcli.sh"
```

The script sets `GOOS`, `GOARCH`, and `CGO_ENABLED=0`, then runs:

```bash
go build -trimpath -buildvcs=false -ldflags="-s -w" -o bin/drawcli-linux-amd64 .
go build -trimpath -buildvcs=false -ldflags="-s -w" -o bin/drawcli-linux-arm64 .
go build -trimpath -buildvcs=false -ldflags="-s -w" -o bin/drawcli-windows-amd64.exe .
go build -trimpath -buildvcs=false -ldflags="-s -w" -o bin/drawcli-darwin-amd64 .
go build -trimpath -buildvcs=false -ldflags="-s -w" -o bin/drawcli-darwin-arm64 .
```

`-s` strips the symbol table and `-w` strips DWARF debug data. Keep both flags
for the compact release artifacts.

## Verify

```bash
SKILL_ROOT="${CLAUDE_SKILL_DIR:-/absolute/path/to/drawsvg}"
cd "$SKILL_ROOT/drawcli"
GOCACHE=/tmp/drawcli-cache go test ./...

file "$SKILL_ROOT/bin/drawcli-linux-amd64" \
     "$SKILL_ROOT/bin/drawcli-linux-arm64" \
     "$SKILL_ROOT/bin/drawcli-windows-amd64.exe" \
     "$SKILL_ROOT/bin/drawcli-darwin-amd64" \
     "$SKILL_ROOT/bin/drawcli-darwin-arm64"

"$SKILL_ROOT/bin/drawcli-linux-amd64" doctor
"$SKILL_ROOT/bin/drawcli-linux-amd64" render memory \
  "$SKILL_ROOT/fixtures/mem0-style1.json" /tmp/drawcli-smoke.svg
"$SKILL_ROOT/bin/drawcli-linux-amd64" check /tmp/drawcli-smoke.svg
```

Expected formats are ELF x86-64/ARM64 for Linux, PE32+ x86-64 for Windows, and
Mach-O x86-64/ARM64 for macOS. Do not claim a release build is valid until
tests, format checks, and the Linux smoke render pass.
