# 04 — Claude Desktop Setup

How to actually plug `jodoo-cli` into Claude Desktop (and by extension Claude
Code, Cursor, Zed — they all read the same MCP config shape).

## 1. Install `jodoo-cli` globally

Claude Desktop spawns the server with the command you give it; if you use a
relative path it will fail. Use an absolute path or put the binary on `PATH`.

```bash
make install         # installs /usr/local/bin/jodoo-cli
which jodoo-cli      # → /usr/local/bin/jodoo-cli
jodoo-cli doctor     # make sure it's green before continuing
```

## 2. Locate `claude_desktop_config.json`

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Linux (Claude Code / CLI variants) | `~/.config/claude/claude_desktop_config.json` |

Create the file if it doesn't exist. It must be valid JSON (no comments, no
trailing commas).

## 3. Minimal config (stdio)

```json
{
  "mcpServers": {
    "jodoo": {
      "command": "/usr/local/bin/jodoo-cli",
      "args": ["mcp", "serve"]
    }
  }
}
```

Restart Claude Desktop. In a new chat:
- Open the **Connectors / Tools** panel
- You should see `jodoo` listed, with the tools enumerated

## 4. Config with a specific profile

If you have multiple Jodoo tenants managed by profiles:

```json
{
  "mcpServers": {
    "jodoo-prod": {
      "command": "/usr/local/bin/jodoo-cli",
      "args": ["--profile", "prod", "mcp", "serve"]
    },
    "jodoo-staging": {
      "command": "/usr/local/bin/jodoo-cli",
      "args": ["--profile", "staging", "mcp", "serve", "--read-only"]
    }
  }
}
```

Note `--read-only` on staging — the server refuses to expose write shortcuts.

## 5. Config with env-based auth

Useful in CI or when you don't want the keychain involved:

```json
{
  "mcpServers": {
    "jodoo": {
      "command": "/usr/local/bin/jodoo-cli",
      "args": ["mcp", "serve"],
      "env": {
        "JODOO_API_KEY": "jd_live_********************",
        "JODOO_OUTPUT":  "json"
      }
    }
  }
}
```

`JODOO_API_KEY` overrides the profile's stored key for the life of that
subprocess. Do not commit this file — it contains the credential.

## 6. Config with logging

Stdio servers can't use stdout for anything except JSON-RPC, so divert our
structured logs somewhere you can tail:

```json
{
  "mcpServers": {
    "jodoo": {
      "command": "/usr/local/bin/jodoo-cli",
      "args": ["mcp", "serve", "--log-file", "/tmp/jodoo-mcp.log"]
    }
  }
}
```

Then:

```bash
tail -f /tmp/jodoo-mcp.log
```

## 7. Scoping tools

Expose only a subset of shortcuts:

```json
{
  "mcpServers": {
    "jodoo-readonly": {
      "command": "/usr/local/bin/jodoo-cli",
      "args": [
        "mcp", "serve",
        "--allow", "jodoo.app-*",
        "--allow", "jodoo.data-list",
        "--allow", "jodoo.data-get",
        "--allow", "jodoo.widget-list"
      ]
    }
  }
}
```

- `--allow` takes glob patterns (`*` matches any segment)
- `--deny` is applied after `--allow`
- With no `--allow`, all tools are exposed (minus `--read-only` filtering and
  `--deny`).

## 8. Verifying the connection

Inside Claude Desktop, paste:

> List the apps reachable by my API key. Use the jodoo connector.

Claude should pick `jodoo.app-list`, call it, and paste the envelope back. If
it instead says "I don't have access to that tool":
- Check the connector is enabled in the UI (Settings → Connectors)
- Check `/tmp/jodoo-mcp.log` for startup errors
- Re-run `jodoo-cli doctor` — auth issues surface here first

## 9. Common pitfalls

| Symptom | Cause | Fix |
|---------|-------|-----|
| Connector never appears | Typo in `claude_desktop_config.json` | Run `jq . < claude_desktop_config.json` |
| "Failed to start server" | `command` not absolute; PATH not inherited | Use `/usr/local/bin/jodoo-cli` explicitly |
| Tool call returns `isError: true` with "api key not configured" | Subprocess doesn't see your keychain | Either set `JODOO_API_KEY` in `env` or run `jodoo-cli config init` and ensure the keychain is unlocked |
| All tool calls hang | Claude Desktop version too old for MCP 2025-06-18 | Update Claude Desktop |
| Server prints warnings to stdout instead of stderr | A shortcut or plumbing function misrouted its output | File a bug — MCP stdio is fragile to stray stdout |

## 10. HTTP transport (remote connector / multi-client)

If you prefer one long-running server backing multiple clients, or need to
run the server on a different host, use the HTTP transport:

```bash
# localhost, auth-gated (recommended even locally)
export MCP_TOKEN=$(openssl rand -hex 32)
jodoo-cli mcp serve --http 127.0.0.1:8765
```

Claude Desktop remote-connector config:

```json
{
  "mcpServers": {
    "jodoo-http": {
      "url": "http://127.0.0.1:8765/mcp",
      "headers": { "Authorization": "Bearer <paste-MCP_TOKEN-here>" }
    }
  }
}
```

Health check: `curl http://127.0.0.1:8765/healthz` → `{"status":"ok"}`.

⚠️ The HTTP transport does **not** do TLS. Bind to loopback
(`127.0.0.1:*`), or put it behind a reverse proxy (nginx, caddy, cloudflared)
that terminates TLS if you need remote access.

## 11. Claude Code / Cursor / Zed

Same `mcpServers` shape, different config location:

- **Claude Code**: `~/.claude/mcp.json`
- **Cursor**: `~/.cursor/mcp.json` (or per-project `.cursor/mcp.json`)
- **Zed**: `~/.config/zed/settings.json` under `"context_servers"`

Translate the key names if needed; the body is the same.
