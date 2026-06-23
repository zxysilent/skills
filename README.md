# Skills

个人的 Agent Skills 集合，用于在日常开发中沉淀可复用的模式与参考指南。

## 目录

| Skill | 说明 |
|-------|------|
| [modern-go](./modern-go/SKILL.md) | Go 代码现代化指南 — 基于 `go.mod` 版本，将遗留写法替换为 modern stdlib API (Go 1.0–1.26) |

## modern-go

**触发条件**：编写或审查 Go 代码时遇到遗留模式（`interface{}`、手动循环、`io/ioutil`、泛型前的手写操作等）。

**核心能力**：
- **55 条 quick lookup 条目**，按版本标注
- **6 个 Phase** 按 Go 版本演进组织代码示例：
  - Phase 1 — Early Cleanups (1.0–1.19)
  - Phase 2 — Generic Renaissance (1.20–1.21)
  - Phase 3 — Syntax & Routing (1.22)
  - Phase 4 — Iterators (1.23)
  - Phase 5 — Quality of Life (1.24)
  - Phase 6 — Present Future (1.25–1.26)
- **12 条 guardrails** 覆盖常见陷阱

**使用方式**：读取项目的 `go.mod` 版本，扫描代码中的遗留模式，仅替换项目版本支持的现代 API。

## 目录结构

```
skills/
  modern-go/
    SKILL.md          # Skill 主文件
  README.md           # 本文件
```

## License

MIT
