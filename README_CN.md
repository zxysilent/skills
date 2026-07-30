# Skills

> [English](./README.md)

个人的 Agent Skills 集合，用于在日常开发中沉淀可复用的模式与参考指南。

## 目录

| Skill | 说明 |
|-------|------|
| [modern-go](./modern-go/SKILL.md) | Go 代码现代化指南 — 基于 `go.mod` 版本，将遗留写法替换为 modern stdlib API (Go 1.0–1.26) |
| [drawsvg](./drawsvg/SKILL.md) | 使用 Go drawcli 生成经过几何校验的 SVG 技术图表 |

## 安装 drawsvg

预编译文件位于项目根目录 [`bins/`](./bins/)。安装时根据当前操作系统和 CPU
架构下载对应二进制到 `drawsvg/scripts/`：

```bash
./drawsvg/scripts/install-drawcli.sh
```

Windows PowerShell 执行：

```powershell
.\drawsvg\scripts\install-drawcli.ps1
```

安装器只下载与平台匹配的文件到 `drawsvg/scripts/`。支持 Linux
amd64/arm64、Windows amd64 以及 macOS amd64/arm64。

## 作为 Codex Skill 安装

在 Codex 环境执行一次以下命令。它只安装 `drawsvg`，再为当前 Linux 或
macOS 平台下载原生运行时：

```bash
SKILLS_HOME="${CODEX_HOME:-$HOME/.codex}/skills"
INSTALLER="$SKILLS_HOME/.system/skill-installer/scripts/install-skill-from-github.py"
PYTHON_BIN="$(command -v python3.12 || command -v python3.11 || command -v python3.10 || command -v python3.9 || command -v python3.8)"

test -n "$PYTHON_BIN" || { echo "需要 Python 3.8+" >&2; exit 1; }
"$PYTHON_BIN" "$INSTALLER" --repo zxysilent/skills --path drawsvg --method git
"$SKILLS_HOME/drawsvg/scripts/install-drawcli.sh"
"$SKILLS_HOME/drawsvg/scripts/drawcli" doctor
```

安装器固定使用 Git，而不是 GitHub archive 下载，以便稳定地完成 sparse
checkout。该方式要求 Python 3.8+；没有兼容 Python 或 Codex 安装器时，使用
下面的手工 fallback：

```bash
SKILLS_HOME="${CODEX_HOME:-$HOME/.codex}/skills"
TEMP_DIR="$(mktemp -d)"
git clone --depth 1 --filter=blob:none --sparse https://github.com/zxysilent/skills.git "$TEMP_DIR"
git -C "$TEMP_DIR" sparse-checkout set drawsvg
mkdir -p "$SKILLS_HOME"
cp -a "$TEMP_DIR/drawsvg" "$SKILLS_HOME/drawsvg"
"$SKILLS_HOME/drawsvg/scripts/install-drawcli.sh"
"$SKILLS_HOME/drawsvg/scripts/drawcli" doctor
```

`doctor` 成功后，先读取 `"$SKILLS_HOME/drawsvg/SKILL.md"` 再使用该 skill。
如果 skill 已存在，跳过复制/安装步骤，只更新匹配平台的 `drawcli` 运行时。
不要直接运行 Go 源码，也不要把其它平台的二进制复制到 `scripts/`。

完整的生成流程、质量检查、风格和图标目录请阅读
[`drawsvg/SKILL.md`](./drawsvg/SKILL.md) 及其 references。

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
  drawsvg/
    SKILL.md          # SVG 图表技能
    scripts/           # 安装器、下载的运行时、构建脚本和 SVG 校验器
  bins/                # 预编译 drawcli 发布文件
  README.md           # 英文说明 (English)
  README_CN.md        # 本文件 (中文)
```

## License

MIT
