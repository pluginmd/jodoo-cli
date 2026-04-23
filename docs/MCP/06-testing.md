# 06 — Testing

Three ways to validate the MCP server, ordered by increasing realism.

## 1. Raw JSON-RPC via heredoc

Fastest sanity check — no Claude involved, no Node, no Python.

```bash
./jodoo-cli mcp serve <<'EOF'
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"jodoo.app-list","arguments":{"limit":5}}}
EOF
```

Expected output (one JSON object per line):

```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-06-18","capabilities":{"tools":{"listChanged":false}},"serverInfo":{"name":"jodoo-cli","version":"…"}}}
{"jsonrpc":"2.0","id":2,"result":{"tools":[…]}}
{"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"{\"code\":0,\"data\":{…}}"}],"isError":false}}
```

If you pipe through `jq -c .` you get one compact object per line for diffing.

## 2. MCP Inspector (recommended)

The official debugging UI for MCP servers. Launches a local web UI, spawns
your server, and gives you a point-and-click way to call tools with form
inputs.

```bash
npx @modelcontextprotocol/inspector /usr/local/bin/jodoo-cli mcp serve
```

Opens `http://localhost:5173` with:
- Protocol log (every frame in both directions)
- Tool catalog (with generated forms from `inputSchema`)
- Live call / response inspector

This is where you'd catch schema regressions quickly — a bad schema shows up
as a broken form.

## 3. End-to-end through Claude Desktop

The final acceptance gate: the actual client.

1. Put the config in `claude_desktop_config.json` (see `04-claude-desktop-setup.md`)
2. Restart Claude Desktop
3. Ask: **"Using the jodoo connector, list my first 5 apps"**
4. Watch Claude call `jodoo.app-list`, approve, see the response
5. Ask: **"Now delete record X from entry Y of app A"**
6. Verify that Claude requires `confirm: true` (it'll say the tool rejected
   without confirmation) — if it goes through on the first try, the risk gate
   is broken.

## Smoke-test checklist

Before tagging a release of the MCP layer:

- [ ] `tools/list` returns every shortcut in the registry (count matches
      `grep -c '^var.*common.Shortcut{' shortcuts/jodoo/*.go`)
- [ ] Every tool has a non-empty `description` and a `type: "object"`
      `inputSchema`
- [ ] `inputSchema.required` matches each flag's `Required: true`
- [ ] `jodoo.data-delete` without `confirm: true` returns `isError: true` with
      an explanatory message
- [ ] `jodoo.data-delete` with `confirm: true` actually calls the API
      (integration; run against a throwaway record)
- [ ] `--read-only` mode hides `jodoo.data-*` write shortcuts
- [ ] `--allow '*.app-*' --deny '*.app-delete'` (if ever added) works
- [ ] `--log-file` receives per-call lines; stderr stays clean of server spam
- [ ] `ping` returns `{}`
- [ ] Sending an unknown method returns JSON-RPC error `-32601`
- [ ] Sending invalid JSON on a line returns `-32700` and the server keeps
      reading subsequent lines (it doesn't die)
- [ ] Canceling a slow tool call from the client (e.g., closing Claude
      Desktop mid-call) kills the forked `jodoo-cli` within 1s

## Unit tests worth writing

These are lightweight and belong under `cmd/mcp/*_test.go`:

- `TestFlagsToSchema_Types` — every `common.Flag.Type` maps to the expected
  JSON Schema shape
- `TestFlagsToSchema_Enum` — `Flag.Enum` becomes `"enum": […]`
- `TestFlagsToSchema_Required` — required flags appear in the `required` list
- `TestBuildArgv_StringArray` — array args become repeated flags
- `TestBuildArgv_Object` — object args are JSON-encoded
- `TestBuildArgv_HighRiskConfirm` — `confirm: true` → `--yes`; missing → error
  before subprocess is forked
- `TestBuildArgv_ReadOnlyBlocksWrite` — write tool not even listed in registry
  under `--read-only`
- `TestDispatch_UnknownMethod` — returns JSON-RPC `-32601`
- `TestDispatch_Notification` — no response for notification-shape messages

No network is required for any of these — they test the adapter layer only.
Real API calls are covered by the existing shortcut integration tests.

## Golden fixture

Once stable, capture `tools/list` output as a fixture and diff-check it in CI:

```bash
./jodoo-cli mcp serve < testdata/tools-list.in.jsonl > testdata/tools-list.out.jsonl
git diff --exit-code testdata/tools-list.out.jsonl
```

This catches accidental breakage of the registry-to-tool mapping when someone
adds a new shortcut flag.
