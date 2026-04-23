# 00 — Overview

## What is MCP?

The **Model Context Protocol** (MCP) is an open JSON-RPC 2.0 protocol that lets
an LLM client (Claude Desktop, Claude Code, Cursor, Zed, Continue, …) talk to
external "servers" that expose three primitives:

| Primitive | Purpose | Used here? |
|-----------|---------|------------|
| **Tools** | Callable functions with a JSON-Schema input + text/image output | Yes — one per shortcut |
| **Resources** | Read-only addressable documents (URIs) | Phase 2 (roadmap) |
| **Prompts** | Parameterised prompt templates | Phase 2 (roadmap) |

Transport is usually **stdio** (the server is a subprocess of the client) or
**streamable HTTP** (the server is a long-running localhost / remote service).
Claude Desktop supports both; stdio is the simplest for a localhost tool like
`jodoo-cli`.

## Why bolt MCP onto `jodoo-cli`?

The CLI is already a carefully curated surface over `api.jodoo.com`:
- ~30 shortcuts with declared input schemas (`common.Flag`)
- Consistent envelope output (`{code, data, msg}` or jq-filtered)
- Built-in `--dry-run`, `--yes` for high-risk writes, `--format`, `--jq`
- Profile-aware auth, keychain-backed credential storage

An MCP layer lets Claude Desktop **reuse every bit of that** without the user
having to remember flag names — Claude pastes them in for you, sees the
structured result, and can chain calls. For write endpoints, the same
`high-risk-write` gate that protects humans also protects the agent.

## What we are NOT building

- ❌ A second parallel command registry (tools duplicated from shortcuts)
- ❌ A separate Python or Node bridge process
- ❌ A custom JSON envelope different from the CLI's
- ❌ Anything that silently bypasses the `--yes` / risk gate

## User-visible result

Before:

```bash
jodoo-cli jodoo +data-list --app-id A --entry-id E --limit 50
```

After, inside Claude Desktop:

> **You:** List the last 50 records in entry `E` of app `A`.
>
> **Claude:** *(calls tool `jodoo.data-list` with `{"app_id":"A","entry_id":"E","limit":50}`)*
> Here are the 50 records …

Behind the scenes the tool handler is exactly the `+data-list` shortcut. Same
validation, same envelope, same error codes.

## What changes in the repo

| File/Dir | Change |
|----------|--------|
| `cmd/mcp/` | **New.** MCP cobra subcommand + stdio server |
| `cmd/root.go` | 1 line to register `mcp` subcommand |
| `docs/MCP/` | **New.** This guide set |
| Everything else | Unchanged. The MCP layer is strictly additive. |

No new Go module dependencies — JSON-RPC 2.0 is ~200 LOC on top of `encoding/json`.
