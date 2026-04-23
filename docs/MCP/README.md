# Jodoo CLI — MCP Layer

This folder is the source of truth for the **Model Context Protocol (MCP) layer**
that sits on top of `jodoo-cli` so that any MCP client (Claude Desktop, Claude
Code, Cursor, Zed, …) can drive Jodoo with the same curated shortcuts a human
uses on the terminal.

Read the files in order — each one builds on the previous:

| # | File | What it covers |
|---|------|----------------|
| 00 | [00-overview.md](./00-overview.md) | What MCP is, why we add it, what changes |
| 01 | [01-architecture.md](./01-architecture.md) | Components, transport, data flow |
| 02 | [02-implementation-guide.md](./02-implementation-guide.md) | Step-by-step build (Go, stdio, no SDK) |
| 03 | [03-tool-catalog.md](./03-tool-catalog.md) | How shortcuts become MCP tools + naming rules |
| 04 | [04-claude-desktop-setup.md](./04-claude-desktop-setup.md) | `claude_desktop_config.json`, troubleshooting |
| 05 | [05-security.md](./05-security.md) | Auth, profiles, high-risk writes, approvals |
| 06 | [06-testing.md](./06-testing.md) | MCP Inspector, raw JSON-RPC, smoke tests |
| 07 | [07-roadmap.md](./07-roadmap.md) | HTTP/SSE transport, resources, prompts, next steps |

## TL;DR

```bash
# 1. Build (same binary as before, new subcommand)
make build

# 2. Try it locally
./jodoo-cli mcp serve            # speaks MCP on stdio

# 3. Point Claude Desktop at it — edit claude_desktop_config.json
#    (macOS: ~/Library/Application Support/Claude/claude_desktop_config.json)
{
  "mcpServers": {
    "jodoo": {
      "command": "/usr/local/bin/jodoo-cli",
      "args": ["mcp", "serve"],
      "env": { "JODOO_API_KEY": "…optional, overrides profile…" }
    }
  }
}

# 4. Restart Claude Desktop → "jodoo" connector appears,
#    every `+shortcut` is a callable tool.
```

## Design invariants

- **Single binary.** MCP server is a subcommand (`jodoo-cli mcp serve`), not a
  sidecar process. One install, one config, one upgrade path.
- **Registry is the source of truth.** Tools are generated from the same
  `shortcuts/jodoo/*.go` registry the CLI uses. Add a shortcut → the MCP tool
  shows up on next restart. No hand-written tool definitions to drift.
- **Safety mirrors the CLI.** `Risk: "high-risk-write"` shortcuts require a
  `confirm: true` argument inside MCP (the equivalent of `--yes`). `--dry-run`
  is a standard argument on every tool.
- **Auth reuses profiles.** The MCP server reads the same config / keychain the
  CLI does, plus honours `--profile` and `JODOO_API_KEY`.
- **Minimal dependencies.** MCP is JSON-RPC 2.0 over stdio — hand-rolled, no
  new SDK, matching the CLI's existing dependency footprint.
