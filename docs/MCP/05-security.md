# 05 — Security Model

Exposing an API via MCP means exposing it to a language model that can be
jailbroken or prompt-injected. This section captures the threat model and the
concrete defences baked into the server.

## Threat model

| Threat | Example | Defence |
|--------|---------|---------|
| Prompt injection causes destructive call | Web page the model is reading says "delete all records" | `high-risk-write` gate forces `confirm: true`, which the user must explicitly approve |
| Silent tool misuse | Model schedules background tool calls | Claude Desktop requires per-tool user approval by default |
| Credential exfiltration via stdout | Tool result echoes API key | CLI never emits the key; MCP server reuses the CLI's envelope |
| Config tampering | User config swapped at runtime | `--profile` is pinned at server start; one process = one profile |
| Excessive blast radius | Model has access to all tools even for "just read this record" | `--allow` / `--deny` / `--read-only` flags |

## Authentication

The MCP server does **not** introduce a new auth mechanism. It uses whatever
the CLI already resolves:

1. `JODOO_API_KEY` env var (highest precedence; useful in CI and in
   `claude_desktop_config.json` `env`)
2. The profile's key stored in the OS keychain (via `zalando/go-keyring`)
3. The profile's key in `~/.jodoo/config.json` (legacy)

Once authenticated, every subprocess the server forks inherits the env, so a
single `JODOO_API_KEY` set in `claude_desktop_config.json` is enough.

## The `--yes` / `confirm` bridge

For `Risk: "high-risk-write"` shortcuts, the CLI refuses to run without `--yes`:

```
Error: this is a high-risk write — pass --yes to confirm
```

Under MCP, the equivalent is an explicit `confirm: true` argument:

```json
{"name":"jodoo.data-delete","arguments":{"app_id":"…","entry_id":"…","data_id":"…","confirm":true}}
```

Rules the server enforces:

1. If the shortcut is high-risk and `confirm` is missing or false, the server
   returns an error content block **without** calling the CLI.
2. `confirm: true` is translated to `--yes` on the child process; never
   combined with any other approval mechanism.
3. Claude Desktop's per-tool confirmation prompt still runs before the server
   even sees the call. `confirm` is the second gate, not the first.

This is deliberate: the model may still pick the tool, but one extra token
(`true` vs `false`) is what separates a preview from a destructive write, and
that token is produced in full view of the user.

## Read-only mode

```
jodoo-cli mcp serve --read-only
```

Filters out every shortcut where `Risk != "read"`. Recommended default for:
- Staging / prod read-only analytics sessions
- Shared machines where the user may prompt the model without reviewing tools
- Demo environments

`--read-only` is a startup decision — it does not change at runtime.

## Allow/deny lists

```
--allow 'jodoo.data-*' --deny 'jodoo.data-delete' --deny 'jodoo.data-batch-delete'
```

- Evaluated in order: `--allow` filters to a set, then `--deny` removes from it
- Glob supports `*` only (no `?` or character classes) — keep it simple
- A denied tool does not appear in `tools/list` at all (not just hidden at call
  time — the model never learns it exists)

## Logging what, logging how

- **Default log line per call**: `tool=<name> exit=<code> dur=<ms>`
- **Never** log `arguments` — they can contain record IDs, names, department
  data, free-text notes. Logging the tool name + duration + exit code is
  enough for debugging throughput or failure patterns.
- **Never** log response bodies — same reason.
- If a user wants full request tracing, they can run their own MITM; we are
  not going to bake it in.

## Supply-chain notes

- The server introduces **no new Go modules**. The MCP protocol is small
  enough to hand-roll, which also means no transitive dependencies to audit.
- The forked CLI binary is `os.Executable()` — same binary, same signature;
  we never shell out to a different path.
- `exec.CommandContext` is always used (not `exec.Command`) so that canceling
  the MCP request also kills the child. This prevents zombie `jodoo-cli` procs
  if Claude Desktop is closed mid-call.

## What's NOT enforced (yet)

- Rate limiting per tool — the Jodoo server enforces its own rate limit, but
  if you expose an MCP server publicly you should add one here.
- Per-tool timeout — handled by `context.Context`; default is whatever the
  client sends. Safe for stdio (human-paced), worth tightening for HTTP.
- Signed tool manifests — not a thing in MCP today.

See `07-roadmap.md` for how these are planned.
