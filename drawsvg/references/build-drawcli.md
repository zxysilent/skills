# Build DrawCLI

`drawcli` is the Go runtime for DrawSVG. Build release binaries from
`drawsvg/drawcli`; do not use the old Python generator as the runtime.

## Bundled Binaries

The repository stores release binaries in its root `bins/` directory:

- `bins/drawcli-linux-amd64`
- `bins/drawcli-linux-arm64`
- `bins/drawcli-windows-amd64.exe`
- `bins/drawcli-darwin-amd64`
- `bins/drawcli-darwin-arm64`

Install the matching binary into `drawsvg/scripts/` from the GitHub repository:

```bash
"$SKILL_ROOT/scripts/install-drawcli.sh"
```

The installer detects OS and CPU architecture, downloads the matching file from
`zxysilent/skills`, saves it as `drawsvg/scripts/drawcli`, and makes it
executable. On Windows, run `scripts/install-drawcli.ps1`; it saves
`scripts/drawcli.exe`.

Override `DRAWCLI_BASE_URL` when using a mirror or a local test server.

## Rebuild Both Platforms

From any working directory:

```bash
SKILL_ROOT="${CLAUDE_SKILL_DIR:-/absolute/path/to/drawsvg}"
GOCACHE=/tmp/drawcli-cache "$SKILL_ROOT/scripts/build-drawcli.sh"
```

The script sets `GOOS`, `GOARCH`, and `CGO_ENABLED=0`, then runs:

```bash
go build -trimpath -buildvcs=false -ldflags="-s -w" -o ../bins/drawcli-linux-amd64 .
go build -trimpath -buildvcs=false -ldflags="-s -w" -o ../bins/drawcli-linux-arm64 .
go build -trimpath -buildvcs=false -ldflags="-s -w" -o ../bins/drawcli-windows-amd64.exe .
go build -trimpath -buildvcs=false -ldflags="-s -w" -o ../bins/drawcli-darwin-amd64 .
go build -trimpath -buildvcs=false -ldflags="-s -w" -o ../bins/drawcli-darwin-arm64 .
```

`-s` strips the symbol table and `-w` strips DWARF debug data. Keep both flags
for the compact release artifacts.

## Verify

```bash
SKILL_ROOT="${CLAUDE_SKILL_DIR:-/absolute/path/to/drawsvg}"
cd "$SKILL_ROOT/drawcli"
GOCACHE=/tmp/drawcli-cache go test ./...

file "$SKILL_ROOT/../bins/drawcli-linux-amd64" \
     "$SKILL_ROOT/../bins/drawcli-linux-arm64" \
     "$SKILL_ROOT/../bins/drawcli-windows-amd64.exe" \
     "$SKILL_ROOT/../bins/drawcli-darwin-amd64" \
     "$SKILL_ROOT/../bins/drawcli-darwin-arm64"

"$SKILL_ROOT/scripts/drawcli" doctor
"$SKILL_ROOT/scripts/drawcli" render memory \
  "$SKILL_ROOT/fixtures/mem0-style1.json" /tmp/drawcli-smoke.svg
"$SKILL_ROOT/scripts/drawcli" check /tmp/drawcli-smoke.svg
```

Expected formats are ELF x86-64/ARM64 for Linux, PE32+ x86-64 for Windows, and
Mach-O x86-64/ARM64 for macOS. Do not claim a release build is valid until
tests, format checks, and the Linux smoke render pass.
