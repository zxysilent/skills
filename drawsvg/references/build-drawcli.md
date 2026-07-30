# Build DrawCLI

`drawcli` is the Go runtime for DrawSVG. Build release binaries from
`drawsvg/drawcli`; do not use the old Python generator as the runtime.

## Bundled Binaries

The repository root stores:

- `bins/drawcli-linux-amd64`
- `bins/drawcli-linux-arm64`
- `bins/drawcli-windows-amd64.exe`
- `bins/drawcli-darwin-amd64`
- `bins/drawcli-darwin-arm64`

The skill installer downloads only the matching artifact into
`drawsvg/scripts/`:

```bash
"$SKILL_ROOT/scripts/install-drawcli.sh"
"$SKILL_ROOT/scripts/drawcli" doctor
```

On Windows, run `scripts/install-drawcli.ps1`; it writes `scripts/drawcli.exe`.
Windows and macOS binaries cannot be smoke-tested on Linux without matching
runners; their target formats can still be checked with `file`.

## Rebuild Release Binaries

From any working directory, use the checked-in build script so every target
uses the same flags and output directory:

```bash
SKILL_ROOT="${CLAUDE_SKILL_DIR:-/absolute/path/to/drawsvg}"
GOCACHE=/tmp/drawcli-cache "$SKILL_ROOT/scripts/build-drawcli.sh"
```

`-s` strips the symbol table and `-w` strips DWARF debug data. Keep both flags
for the compact release artifacts. The script sets `GOOS`, `GOARCH`, and
`CGO_ENABLED=0` for Linux amd64/arm64, Windows amd64, and macOS amd64/arm64.

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

"$SKILL_ROOT/scripts/install-drawcli.sh"
"$SKILL_ROOT/scripts/drawcli" doctor
"$SKILL_ROOT/scripts/drawcli" render memory \
  "$SKILL_ROOT/fixtures/mem0-style1.json" /tmp/drawcli-smoke.svg
"$SKILL_ROOT/scripts/drawcli" check /tmp/drawcli-smoke.svg
```

Expected formats are ELF x86-64/ARM64 for Linux, PE32+ x86-64 for Windows, and
Mach-O x86-64/ARM64 for macOS. Do not claim a release build is valid until
tests, format checks, and the Linux smoke render pass.
