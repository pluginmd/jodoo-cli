# 01 — Architecture

## Component map

```
┌───────────────────────────────────────────────────────────┐
│                    Claude Desktop                         │
│  (MCP client, spawns jodoo-cli as stdio subprocess)       │
└───────────────┬───────────────────────────────┬───────────┘
                │ JSON-RPC 2.0 over stdio       │
                │  (newline-delimited frames)   │
                ▼                               ▲
┌───────────────────────────────────────────────────────────┐
│                jodoo-cli mcp serve                        │
│                                                           │
│  cmd/mcp/server.go   ← stdio reader/writer loop           │
│  cmd/mcp/tools.go    ← tools/list, tools/call dispatch    │
│  cmd/mcp/schema.go   ← Flag[] → JSON Schema               │
│  cmd/mcp/dispatch.go ← invoke shortcut via subprocess     │
│                                                           │
│             uses the same registry ▼                      │
│                                                           │
│  shortcuts/jodoo/*  (app, data, file, workflow, contact)  │
│  shortcuts/common/  (RuntimeContext, Flag, Shortcut)      │
│                                                           │
│             uses the same plumbing ▼                      │
│                                                           │
│  internal/client   ← HTTP client, auth header             │
│  internal/core     ← profile / config resolution          │
│  internal/output   ← envelope, ExitError, jq              │
└───────────────────────────────────────────────────────────┘
                │ HTTPS POST
                ▼
            api.jodoo.com
```

## Transport: why stdio first

Claude Desktop supports three transports for MCP servers:

| Transport | Use case | Status here |
|-----------|----------|-------------|
| **stdio** | Local binary, 1 client, simplest | ✅ default |
| **Streamable HTTP** | Remote / multi-client servers | ✅ `--http <addr>` (SSE streaming not yet) |
| **SSE** | Deprecated in favour of streamable HTTP | Not supported |

Stdio is the right default because:
1. No port to negotiate, no TLS, no auth token for the transport itself.
2. Lifecycle is tied to Claude Desktop — the subprocess dies with the client.
3. Matches how every other local MCP server ships (`filesystem`, `postgres`, `git`).

HTTP is useful if you want one long-running server to back multiple clients,
or to run it on a remote host. See `07-roadmap.md`.

## Dispatch strategy: fork-self

A call `tools/call { name: "jodoo.data-list", arguments: {…} }` does **not**
import and run `Shortcut.Execute` in-process. Instead it:

1. Looks up the shortcut by name.
2. Translates `arguments` (a JSON object) into argv: `jodoo-cli jodoo +data-list --app-id … --entry-id …`.
3. Applies profile / env overrides, adds `--format json` (or `--dry-run` if requested).
4. `exec.Cmd.Run()` a child `jodoo-cli` process.
5. Captures stdout (the envelope) and stderr (error diagnostics).
6. Wraps the stdout JSON as an MCP `text` content block, or produces an error
   result with the non-zero exit code and stderr.

Why fork instead of in-process dispatch?

- **Identical behaviour to the CLI.** If `+data-list` works on the terminal,
  it works in MCP — no "why does it behave differently in Claude?" debugging.
- **Isolation.** A shortcut panicking, mutating flags, writing to `os.Stdout`,
  or leaking goroutines cannot corrupt the MCP server. The process boundary is
  the cheapest possible sandbox.
- **Concurrency.** MCP tool calls are allowed to arrive concurrently. The
  existing shortcut runner uses shared `cobra.Command` flag storage, which is
  not safe to reuse concurrently. One child per call = no shared state.
- **Simpler security gating.** The `high-risk-write` flag already lives on the
  CLI; the MCP layer just has to pass or withhold `--yes`.

Cost: ~5–20 ms per call for the process fork, plus JSON parse of CLI stdout.
With a human-in-the-loop client that is invisible. If we ever needed raw
throughput we would switch to in-process dispatch (noted in the roadmap).

## JSON-RPC framing

MCP on stdio uses **newline-delimited JSON** (one object per line), not the
HTTP-style `Content-Length:` framing used by LSP. Each frame is a standard
JSON-RPC 2.0 Request, Response, or Notification:

```json
// client → server
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{…}}

// server → client
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-06-18","capabilities":{"tools":{}},"serverInfo":{…}}}

// client → server (notification, no id, no response)
{"jsonrpc":"2.0","method":"notifications/initialized"}
```

Required server methods for Phase 1:

| Method | Purpose |
|--------|---------|
| `initialize` | Handshake, declares capabilities |
| `notifications/initialized` | Client ack, no response |
| `tools/list` | Return every shortcut as a tool definition |
| `tools/call` | Execute a tool, return result |
| `ping` | Health check |
| `shutdown` *(optional)* | Graceful stop |

Nothing else is required to ship a usable connector.

## File ownership map

| File | Owns |
|------|------|
| `cmd/mcp/mcp.go` | Cobra subcommand + CLI flags (`--http`, `--log-file`, `--allow`, `--deny`) |
| `cmd/mcp/server.go` | JSON-RPC read/write loop, dispatch table, error envelope |
| `cmd/mcp/tools.go` | `tools/list` + `tools/call` implementation |
| `cmd/mcp/schema.go` | `common.Flag[] → JSONSchema` translation |
| `cmd/mcp/dispatch.go` | Build argv, fork `jodoo-cli`, wrap stdout/stderr |
| `cmd/mcp/log.go` | Structured logging to `--log-file` (stderr is reserved by MCP) |
| `docs/MCP/*.md` | This guide set |

`cmd/root.go` gains exactly one line: `rootCmd.AddCommand(mcp.NewCmdMcp(f))`.
