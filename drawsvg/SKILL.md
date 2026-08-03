---
name: drawsvg
description: >-
  Use when a user asks to visualize a software or engineering concept as a
  technical diagram, including architecture, data flow, agents, memory,
  sequence, UML, cloud, event-stream, observability, or network topology.
  Use for SVG diagrams only; do not use for photos, raster artwork, or data
  charts.
---

# DrawSVG

Generate geometry-checked technical diagrams from JSON and export SVG. The
only runtime is the bundled Go `drawcli`; SVG is the only supported export.

## Runtime

Resolve the directory containing this file as `SKILL_ROOT` before reading files
or running commands. Do not rely on a working-directory change persisting
between shell calls.

- Codex: use the absolute skill directory from the loaded skill metadata.
- Claude Code: use `${CLAUDE_SKILL_DIR}`.

The release binaries, installation, and cross-compilation instructions for
Linux, Windows, and macOS are in
[`references/build-drawcli.md`](references/build-drawcli.md).

The runtime is provider-neutral by default. Style 10 uses the built-in icon
catalog documented in [`references/style-10-cloud-fabric.md`](references/style-10-cloud-fabric.md);
vendor logos are external adapters and are not bundled.

## Workflow

1. Classify the diagram and load [`references/diagram-types-and-rules.md`](references/diagram-types-and-rules.md).
2. Extract nodes, containers, layers, edges, flows, and semantic groups.
3. Load [`references/composition-quality-contract.md`](references/composition-quality-contract.md) and plan a collision-free layout.
4. Select Style 1 by default; load the matching `references/style-N-*.md` for exact tokens.
5. For Styles 9-12, select the matching semantic contract and validate before rendering.
6. Map concepts to shapes and icons using `diagram-types-and-rules.md` and `references/icons.md`.
7. Write the JSON payload and render with `drawcli`.
8. Run SVG checks, rasterize representative outputs when possible, and visually inspect them.

## Commands

```bash
SKILL_ROOT="${CLAUDE_SKILL_DIR:-/absolute/path/to/drawsvg}"

"$SKILL_ROOT/scripts/install-drawcli.sh"
DRAWCLI="$SKILL_ROOT/scripts/drawcli"

"$DRAWCLI" validate architecture diagram.json
"$DRAWCLI" render architecture diagram.json out.svg --report layout.json
"$DRAWCLI" check out.svg
"$DRAWCLI" inspect out.svg
"$SKILL_ROOT/scripts/validate-svg.sh" out.svg
```

On Windows, run `scripts/install-drawcli.ps1` from PowerShell; it installs
`scripts/drawcli.exe`. On Unix, the installer detects the host OS/CPU and
installs `scripts/drawcli`.

The release binaries are kept in the repository root `bins/` directory. They
are not copied into the skill directory; installation downloads only the
matching artifact into `scripts/`.

After installation, render with the installed binary:

```bash
"$SKILL_ROOT/scripts/drawcli" render architecture diagram.json out.svg
```

`drawcli` renders the JSON payloads used by the bundled fixtures. The `mode`
argument must agree with the payload's template type.

Style 8 is a hand-authored static SVG and cannot be generated from a template;
use the bundled `fixtures/dark-luxury-style8.svg` as the source instead of
rendering it with `drawcli`.

## Quality Gates

- Keep a `viewBox`, explicit node ports, resolved marker URLs, escaped text,
  and painter order: background, containers, edges, nodes, labels, legend.
- Keep at least 40px canvas margins and 60px between unrelated nodes; avoid
  routes through nodes, titles, legends, or footers.
- Prefer orthogonal routes with at most two bends on showcase diagrams.
- Keep labels inside their parent region with readable clearance; use explicit
  route fields only when ordinary routing cannot satisfy the layout contract.
- Always include a legend when multiple flow semantics are present.
- Do not claim visual correctness without rasterizing and inspecting an output.

Before delivery, run:

```bash
"$SKILL_ROOT/scripts/drawcli" validate <mode> diagram.json
"$SKILL_ROOT/scripts/drawcli" render <mode> diagram.json out.svg --report layout.json
"$SKILL_ROOT/scripts/drawcli" check out.svg
```

For arrow-only fixes, preserve nodes, containers, style, and layout; change
only the relevant `arrows` entries, then rerender and recheck.

## Style Selection

| Style | Name | Typical use |
|---:|---|---|
| 1 | Flat Icon (default) | General architecture and memory diagrams |
| 2 | Dark Terminal | Developer and code-oriented diagrams |
| 3 | Blueprint | Formal architecture and infrastructure docs |
| 4 | Notion Clean | Minimal operational documentation |
| 5 | Glassmorphism | Product and presentation diagrams |
| 6 | Claude Official | Warm editorial documentation |
| 7 | OpenAI | Product and API diagrams |
| 8 | Dark Luxury | Hand-authored premium SVG only |
| 9 | C4 Review Canvas | C4 and ADR reviews |
| 10 | Cloud Fabric | Cloud ownership and deployment |
| 11 | Event Transit | Topics, processors, and DLQs |
| 12 | Ops Pulse | Golden signals and traces |

Load style references on demand; do not duplicate their color tokens in this
file.

## Common Topologies

- RAG: Query -> Embed -> Vector Search -> Retrieve -> Augment -> LLM -> Response
- Agentic RAG: Query -> Planner -> Tool use -> LLM, with an explicit loop
- Agent memory: Input -> Memory Manager -> Write/Read stores -> Context
- Multi-agent: Orchestrator -> Subagents -> Aggregator -> Output
- Tool call: LLM -> Tool Selector -> Tool Execution -> Result Parser -> LLM

## References

| Topic | Reference |
|---|---|
| Diagram types, UML, shapes, arrows, validation | `references/diagram-types-and-rules.md` |
| Composition and spacing budgets | `references/composition-quality-contract.md` |
| Product icons | `references/icons.md` |
| Cloud icon catalog and provider boundary | `references/style-10-cloud-fabric.md` |
| Style selection matrix | `references/style-diagram-matrix.md` |
| Style tokens | `references/style-N-*.md` |
| Go build and release binaries | `references/build-drawcli.md` |
