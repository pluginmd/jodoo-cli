# `jodoo-cli` — Docs

Design thinking, architecture, and reference material for contributors
and agents working on `jodoo-cli`.

## Contributor docs

| File | Reading order | What it covers |
|---|---|---|
| [`project-overview.md`](./project-overview.md) | 1 | Scope, goals, non-goals, audience, success criteria. |
| [`design-principles.md`](./design-principles.md) | 2 | The *why* behind each major decision (declarative shortcuts, escape hatch, error shape, credential strategy, …). |
| [`system-architecture.md`](./system-architecture.md) | 3 | Top-level diagram, package responsibilities, request lifecycles, config resolution chain. |
| [`codebase-map.md`](./codebase-map.md) | 4 | File-by-file tour. Use as a lookup when `grep` is too coarse. |
| [`adding-a-shortcut.md`](./adding-a-shortcut.md) | 5 | Step-by-step cookbook for wrapping a new Jodoo endpoint. |
| [`code-standards.md`](./code-standards.md) | 6 | Conventions: naming, errors, tests, commits, dependencies. |

## API reference

| Directory | Purpose |
|---|---|
| [`jodoo-sdk-docs/`](./jodoo-sdk-docs/) | Authoritative Jodoo API specs split into 8 files (developer guide, app, form/data, file, workflow, contact, webhook, error codes). Source of truth for every shortcut implementation. |
| [`research.md`](./research.md) | Short index pointing into `jodoo-sdk-docs/`. |

## Integration layers

| Directory | Purpose |
|---|---|
| [`MCP/`](./MCP/) | Model Context Protocol server that exposes every shortcut to Claude Desktop / Claude Code / Cursor / Zed. 9 files: overview, architecture, implementation guide, tool catalog, Claude Desktop setup, security, testing, roadmap. |

## Reading paths by role

- **Just want to use the CLI?** Start at the top-level
  [`README.md`](../README.md) and the cheat-sheet in
  [`skills/working/jodoo-cli/SKILL.md`](../skills/working/jodoo-cli/SKILL.md).
- **New contributor?** `project-overview` →
  `design-principles` → `adding-a-shortcut`.
- **Onboarding an agent / LLM?** The skill files under
  `skills/working/` are the canonical brief. This docs tree is the
  extended reference behind them.
- **Reviewing a shortcut PR?** Open `adding-a-shortcut.md` and
  `jodoo-sdk-docs/` side by side; the PR should match both.
- **Plugging the CLI into Claude Desktop (or any MCP client)?** Read
  [`MCP/README.md`](./MCP/README.md) — the guide ships with a working
  `jodoo-cli mcp serve` implementation.
