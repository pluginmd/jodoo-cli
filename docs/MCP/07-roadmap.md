# 07 — Roadmap

Phase 1 (this folder) ships stdio + tools. Everything below is explicitly out
of scope for the first cut but ordered by likely usefulness.

## Phase 2 — HTTP transport ✅ SHIPPED

HTTP transport is implemented. See [`cmd/mcp/http.go`](../../cmd/mcp/http.go).

Endpoints:

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/mcp` | Bearer (if `--token`) | JSON-RPC request → JSON-RPC response |
| GET | `/mcp` | Bearer | 405 — SSE stream not yet implemented |
| GET | `/healthz` | open | liveness probe |
| GET | `/info` | open | server metadata (no tool list) |

Run:

```bash
# open, localhost only (no auth)
jodoo-cli mcp serve --http 127.0.0.1:8765

# bearer-token auth via flag
jodoo-cli mcp serve --http 127.0.0.1:8765 --token "$(openssl rand -hex 32)"

# bearer-token via env (preferred — flag leaks into ps output)
MCP_TOKEN=$(openssl rand -hex 32) jodoo-cli mcp serve --http 127.0.0.1:8765
```

Claude Desktop remote-connector config:

```json
{
  "mcpServers": {
    "jodoo": {
      "url": "http://127.0.0.1:8765/mcp",
      "headers": { "Authorization": "Bearer <same-token-as-server>" }
    }
  }
}
```

Still TODO (phase 2.5):
- Full Streamable-HTTP SSE (server-to-client stream for progress
  notifications during long tool calls)
- `Mcp-Session-Id` header for multi-request session tracking
- TLS termination — run behind a reverse proxy today (nginx, caddy, cloudflared)

## Phase 3 — Resources

MCP resources are read-only, addressable documents. They let the model fetch
context without a tool call.

Candidate URIs:

| URI | Content |
|-----|---------|
| `jodoo://app/{app_id}` | App metadata (name, entries) |
| `jodoo://app/{app_id}/entry/{entry_id}/schema` | Widget list (form schema) |
| `jodoo://app/{app_id}/entry/{entry_id}/record/{data_id}` | Single record |
| `jodoo://contact/member/{username}` | Member profile |

Exposing these means the model can say "fetch `jodoo://app/X/entry/Y/schema`"
instead of calling `jodoo.widget-list` — it's lighter-weight and cacheable on
the client side.

Implementation: add `resources/list` and `resources/read` handlers that re-use
the read shortcuts.

## Phase 4 — Prompts

MCP prompts are parameterised prompt templates surfaced to the client UI.
They show up in Claude Desktop as slash-command-style shortcuts.

Candidates:

| Name | Args | Body |
|------|------|------|
| `summarise-records` | `app_id, entry_id, limit` | "Summarise the last {limit} records in {entry_id}…" |
| `find-approval-backlog` | `username` | "List tasks assigned to {username} pending more than 3 days" |
| `audit-high-risk` | `since` | "List every high-risk write that executed since {since}" |

These are small but make the product feel a lot more "integrated" than a
bag-of-tools.

## Phase 5 — In-process dispatch

Only worth doing if HTTP throughput becomes a real constraint. Replaces
fork-self with direct `Shortcut.Execute` invocation, using a per-call cloned
`cobra.Command` to avoid flag-state races.

Blockers before doing this:
- Audit `shortcuts/jodoo/*` for any code that writes directly to
  `os.Stdout` / `os.Stderr` (bypassing `RuntimeContext.IO()`). The output
  plumbing is already mostly clean but hasn't been stressed with concurrent
  goroutines.
- Confirm the HTTP client is safe to share across goroutines (it is —
  `net/http.Client` is goroutine-safe, but the current wrapper hasn't been
  proven under concurrent use).

Estimated gain: ~10ms per call, saved 30-40% overall wall time for very small
tool payloads. For most real workloads the API latency dominates.

## Phase 6 — Progress notifications

For long-running tools (`+member-batch-import`, `+department-batch-import`,
paginating large datasets), emit `notifications/progress` so the client can
show a spinner with real progress instead of hanging.

Requires: the shortcut has to produce intermediate output. Today they don't —
they block until the whole pagination is done. A minimal change: emit
`notifications/progress` per page from the dispatch layer by watching child
stdout and looking for the existing pagination log lines.

## Phase 7 — Tool manifest signing / trust

Once the ecosystem has standardised on signed manifests (nothing concrete
yet), add signature verification to Claude Desktop config so a tampered
binary can't pose as the jodoo connector.

## Phase 8 — Rate limiting + quota

Only matters if the server is multi-tenant (HTTP mode). Options:
- Token-bucket per tool (protect expensive ones like `+data-list` over large
  entries)
- Per-connection quota (protect the Jodoo API quota itself)

Out of scope until someone asks for it.

## Non-goals

- **Multi-model adapters.** This server is MCP; other protocols belong in
  different binaries.
- **UI for picking tools.** That's Claude Desktop's job.
- **Persistent state across calls.** Each `tools/call` is a fresh subprocess;
  that's intentional.
