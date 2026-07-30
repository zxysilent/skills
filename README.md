# Skills

> [中文说明](./README_CN.md)

A personal collection of Agent Skills — documenting reusable patterns and
reference guides accumulated through daily development.

## Catalog

| Skill | Description |
|-------|-------------|
| [modern-go](./modern-go/SKILL.md) | Modernizing Go Code — replace legacy idioms with modern stdlib APIs based on the project's `go.mod` version (Go 1.0–1.26) |
| [drawsvg](./drawsvg/SKILL.md) | Generate geometry-checked technical diagrams as SVG with the Go drawcli runtime |

## Install drawsvg

The release binaries are stored in the repository root [`bins/`](./bins/).
Install the binary matching the current operating system and CPU architecture
into `drawsvg/scripts/`:

```bash
./drawsvg/scripts/install-drawcli.sh
```

On Windows PowerShell:

```powershell
.\drawsvg\scripts\install-drawcli.ps1
```

The installer downloads only the matching binary from `bins/` into
`drawsvg/scripts/`. Linux amd64/arm64, Windows amd64, and macOS amd64/arm64
are supported.

## Autonomous Agent Installation

When an agent needs this skill and the repository is not already installed,
use the host skill installer first. For Codex, run:

```bash
SKILLS_HOME="${CODEX_HOME:-$HOME/.codex}/skills"
python3 "$SKILLS_HOME/.system/skill-installer/scripts/install-skill-from-github.py" \
  --repo zxysilent/skills --path drawsvg

SKILL_ROOT="$SKILLS_HOME/drawsvg"
"$SKILL_ROOT/scripts/install-drawcli.sh"
"$SKILL_ROOT/scripts/drawcli" doctor
```

After `doctor` succeeds, read `"$SKILL_ROOT/SKILL.md"` and follow its
workflow. If `drawsvg` already exists in the agent skill directory, skip the
GitHub install step, inspect its `SKILL.md`, and only install or refresh the
matching `drawcli` runtime. Do not run the Go source directly or copy a
different platform binary into `scripts/`.

For the complete generation workflow, quality gates, styles, and icon catalog,
read [`drawsvg/SKILL.md`](./drawsvg/SKILL.md) and its linked references.

## modern-go

**Trigger**: Writing or reviewing Go code and encountering legacy patterns
(`interface{}`, manual loops, `io/ioutil`, pre-generics hand-rolled operations,
etc.).

**Capabilities**:
- **55 quick-lookup entries**, annotated with Go version requirements
- **6 Phases** organized by Go release timeline with code examples:
  - Phase 1 — Early Cleanups (1.0–1.19)
  - Phase 2 — Generic Renaissance (1.20–1.21)
  - Phase 3 — Syntax & Routing (1.22)
  - Phase 4 — Iterators (1.23)
  - Phase 5 — Quality of Life (1.24)
  - Phase 6 — Present Future (1.25–1.26)
- **12 guardrails** covering common edge cases and pitfalls

**How it works**: Reads the project's `go.mod` version, scans code for legacy
patterns, and only suggests replacements supported by that version.

## Directory Structure

```
skills/
  modern-go/
    SKILL.md          # Skill definition
  drawsvg/
    SKILL.md          # SVG diagram skill
    scripts/           # Installer, downloaded runtime, build script, and SVG checker
  bins/                # Precompiled drawcli releases
  README.md           # This file (English)
  README_CN.md        # 中文说明 (Chinese)
```

## License

MIT
