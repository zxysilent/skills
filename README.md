# Skills

> [中文说明](./README_CN.md)

A personal collection of Agent Skills — documenting reusable patterns and
reference guides accumulated through daily development.

## Catalog

| Skill | Description |
|-------|-------------|
| [modern-go](./modern-go/SKILL.md) | Modernizing Go Code — replace legacy idioms with modern stdlib APIs based on the project's `go.mod` version (Go 1.0–1.26) |

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
  README.md           # This file (English)
  README_CN.md        # 中文说明 (Chinese)
```

## License

MIT
