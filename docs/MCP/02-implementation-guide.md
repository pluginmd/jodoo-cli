# 02 — Implementation Guide

This file walks through every file in `cmd/mcp/` so you can re-derive the
implementation from scratch if it ever gets lost. Code shown is the canonical
version — keep this doc and the code in lockstep.

## 0. Prereqs

- Go ≥ 1.23 (same as the rest of the repo)
- `make build` works (produces `./jodoo-cli`)
- You have a working Jodoo API key (`jodoo-cli doctor` green)

No new modules needed. `encoding/json`, `bufio`, `os/exec`, `context` are enough.

## 1. Cobra subcommand skeleton — `cmd/mcp/mcp.go`

Responsibility: parse MCP-server-specific flags, hand off to `Serve()`.

Flags:

| Flag | Type | Default | Purpose |
|------|------|---------|---------|
| `--http` | string | "" | Bind address (e.g. `:8765`). Empty = stdio. Phase 2. |
| `--log-file` | string | "" | Write structured logs here; stderr reserved for MCP-required diagnostics |
| `--allow` | []string | nil | Optional allow-list of tool names (glob allowed) |
| `--deny` | []string | nil | Optional deny-list (applied after allow) |
| `--read-only` | bool | false | Refuse any shortcut whose `Risk != "read"` |

```go
package mcp

import (
    "github.com/spf13/cobra"
    "jodoo-cli/internal/cmdutil"
)

func NewCmdMcp(f *cmdutil.Factory) *cobra.Command {
    c := &cobra.Command{Use: "mcp", Short: "Model Context Protocol layer"}
    c.AddCommand(newCmdServe(f))
    return c
}

func newCmdServe(f *cmdutil.Factory) *cobra.Command {
    opts := &ServeOptions{}
    c := &cobra.Command{
        Use:   "serve",
        Short: "Run an MCP server over stdio (default) or HTTP",
        RunE:  func(cmd *cobra.Command, _ []string) error { return Serve(cmd.Context(), f, opts) },
    }
    c.Flags().StringVar(&opts.HTTPAddr, "http", "", "bind HTTP server (empty = stdio)")
    c.Flags().StringVar(&opts.LogFile, "log-file", "", "write diagnostic logs here")
    c.Flags().StringSliceVar(&opts.Allow, "allow", nil, "allow-list of tool names (glob)")
    c.Flags().StringSliceVar(&opts.Deny, "deny", nil, "deny-list of tool names (glob)")
    c.Flags().BoolVar(&opts.ReadOnly, "read-only", false, "reject non-read shortcuts")
    return c
}
```

## 2. Server loop — `cmd/mcp/server.go`

Stdio frames = **newline-delimited JSON** (not LSP framing). Read a line, parse
as JSON-RPC Request, dispatch, write the response on one line.

Pseudo-code:

```go
func Serve(ctx context.Context, f *cmdutil.Factory, opts *ServeOptions) error {
    if opts.HTTPAddr != "" { return serveHTTP(ctx, f, opts) }   // phase 2
    return serveStdio(ctx, f, opts)
}

func serveStdio(ctx context.Context, f *cmdutil.Factory, opts *ServeOptions) error {
    reg := buildRegistry(f, opts)           // tools.go
    in  := bufio.NewScanner(os.Stdin)
    in.Buffer(make([]byte, 1<<20), 64<<20)  // allow large tool results
    out := json.NewEncoder(os.Stdout)
    var mu sync.Mutex                       // serialise Encode() calls

    for in.Scan() {
        line := in.Bytes()
        if len(bytes.TrimSpace(line)) == 0 { continue }

        var req rpcRequest
        if err := json.Unmarshal(line, &req); err != nil {
            writeErr(out, &mu, nil, -32700, "parse error", err.Error())
            continue
        }

        go func(req rpcRequest) {           // concurrent tool calls OK
            resp := dispatch(ctx, reg, &req)
            if req.ID == nil { return }     // notification → no response
            mu.Lock(); defer mu.Unlock()
            _ = out.Encode(resp)
        }(req)
    }
    return in.Err()
}
```

`dispatch` is a small router:

```go
func dispatch(ctx context.Context, reg *Registry, req *rpcRequest) *rpcResponse {
    switch req.Method {
    case "initialize":                return handleInitialize(req)
    case "notifications/initialized": return nil
    case "ping":                      return ok(req, struct{}{})
    case "tools/list":                return handleToolsList(reg, req)
    case "tools/call":                return handleToolsCall(ctx, reg, req)
    case "shutdown":                  return ok(req, struct{}{})
    default:                          return rpcErr(req, -32601, "method not found", req.Method)
    }
}
```

`handleInitialize` returns the canonical capability blob:

```json
{
  "protocolVersion": "2025-06-18",
  "capabilities": { "tools": { "listChanged": false } },
  "serverInfo": { "name": "jodoo-cli", "version": "<build.Version>" }
}
```

## 3. Schema generation — `cmd/mcp/schema.go`

Each MCP tool must declare `inputSchema` as a JSON Schema `{"type":"object",…}`.
We generate this from `common.Flag[]`:

| `Flag.Type` | JSON Schema |
|-------------|-------------|
| `""` (string) | `{"type":"string"}` |
| `bool` | `{"type":"boolean"}` |
| `int` | `{"type":"integer"}` |
| `string_array` | `{"type":"array","items":{"type":"string"}}` |

Plus:
- `Flag.Enum` → `"enum": […]`
- `Flag.Required` → added to top-level `required` array
- `Flag.Desc` → `"description"`
- Input hint (`@file`, `-` stdin) appended to description; not expressible in schema

We also add two synthetic arguments on every tool:
- `dry_run: boolean` — mapped to `--dry-run`
- `confirm: boolean` — only for `Risk == "high-risk-write"`, mapped to `--yes`

```go
func flagsToSchema(s common.Shortcut) map[string]interface{} {
    props := map[string]interface{}{}
    required := []string{}
    for _, fl := range s.Flags {
        prop := flagToProp(fl)
        props[flagKey(fl.Name)] = prop
        if fl.Required { required = append(required, flagKey(fl.Name)) }
    }
    props["dry_run"] = map[string]interface{}{"type":"boolean","description":"preview request without executing"}
    if s.Risk == "high-risk-write" {
        props["confirm"] = map[string]interface{}{"type":"boolean","description":"must be true for high-risk writes"}
    }
    schema := map[string]interface{}{"type":"object","properties":props,"additionalProperties":false}
    if len(required) > 0 { schema["required"] = required }
    return schema
}
```

`flagKey` converts `app-id` → `app_id` (MCP convention tends to snake_case and
Claude handles it slightly better than kebab-case, though both work).

## 4. Tool registry — `cmd/mcp/tools.go`

```go
type Tool struct {
    Name        string                 // "jodoo.data-list"
    Description string                 // human-readable
    Risk        string                 // carried from shortcut
    InputSchema map[string]interface{} // generated
    sc          common.Shortcut        // underlying shortcut (private)
}

type Registry struct {
    factory *cmdutil.Factory
    opts    *ServeOptions
    tools   map[string]*Tool
    order   []string
}

func buildRegistry(f *cmdutil.Factory, opts *ServeOptions) *Registry {
    reg := &Registry{factory: f, opts: opts, tools: map[string]*Tool{}}
    for _, p := range shortcutProviders() {        // shortcuts.providers()
        for _, sc := range p.All() {
            name := p.Service() + "." + strings.TrimPrefix(sc.Command, "+")
            if opts.ReadOnly && sc.Risk != "" && sc.Risk != "read" { continue }
            if !matchAllow(name, opts.Allow, opts.Deny)        { continue }
            reg.tools[name] = &Tool{
                Name: name, Description: sc.Description, Risk: sc.Risk,
                InputSchema: flagsToSchema(sc), sc: sc,
            }
            reg.order = append(reg.order, name)
        }
    }
    sort.Strings(reg.order)
    return reg
}
```

> `shortcutProviders()` re-uses the existing `shortcuts` package. We export a
> thin helper there (or access the provider interface directly) so we don't
> hand-duplicate the list.

`tools/list` response shape (per MCP spec):

```json
{
  "tools": [
    {
      "name": "jodoo.data-list",
      "description": "List records in a form with optional filtering",
      "inputSchema": {"type":"object","properties":{…},"required":["app_id","entry_id"]}
    },
    ...
  ]
}
```

## 5. Dispatch (fork-self) — `cmd/mcp/dispatch.go`

```go
func (t *Tool) Call(ctx context.Context, args map[string]interface{}, opts *ServeOptions) (*toolResult, error) {
    argv, err := buildArgv(t, args, opts)           // snake_case→kebab-case, json encode complex values
    if err != nil { return nil, err }

    self, err := os.Executable()
    if err != nil { return nil, err }

    cmd := exec.CommandContext(ctx, self, argv...)
    cmd.Env  = append(os.Environ(), "JODOO_OUTPUT=json")
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr

    err = cmd.Run()
    return &toolResult{
        isError: err != nil,
        stdout:  stdout.Bytes(),
        stderr:  stderr.Bytes(),
        exit:    cmd.ProcessState.ExitCode(),
    }, nil
}
```

`buildArgv` turns `{"app_id":"A","entry_id":"E","limit":50,"dry_run":true}` into:

```
["jodoo", "+data-list", "--app-id", "A", "--entry-id", "E", "--limit", "50", "--dry-run"]
```

Rules:
- Bool true → `--flag` with no value; bool false → omit.
- Arrays → repeated `--flag value` (matches `StringArray`).
- Objects → JSON-encoded, passed as the string value (matches `--data '{...}'`).
- `dry_run` → `--dry-run` (never `--dry-run=false`).
- `confirm: true` + high-risk → `--yes`. Otherwise strip.
- `--format json` is appended for every tool that has `HasFormat=true`, unless
  `dry_run` is true (dry-run response is JSON on stdout already).

The MCP response wraps the CLI stdout as content:

```json
{
  "content": [ { "type": "text", "text": "<CLI stdout — the envelope>" } ],
  "isError": false
}
```

If the CLI exits non-zero:

```json
{
  "content": [ { "type": "text", "text": "<stderr>" } ],
  "isError": true
}
```

> `isError: true` is semantic (the model sees that the call failed), not a
> JSON-RPC error. A JSON-RPC error is only emitted for protocol-level
> failures (unknown method, malformed input).

## 6. Logging — `cmd/mcp/log.go`

Important constraint on stdio: **stdout is JSON-RPC, stderr is free-form
diagnostics consumed by Claude Desktop's log viewer**. Never write anything to
stdout except the JSON-RPC frames. Our own structured logs go to:

1. `--log-file <path>` if provided, or
2. stderr (Claude Desktop will capture it), tagged `[jodoo-mcp]`.

Minimum lines to log:
- Server start + protocol version negotiated
- Every `tools/call` with tool name, duration, exit code (never log `arguments` — they may contain PII/secrets)
- Every protocol error

## 7. Wiring into root — `cmd/root.go`

One line next to the other `AddCommand` calls:

```go
import jodoomcp "jodoo-cli/cmd/mcp"
…
rootCmd.AddCommand(jodoomcp.NewCmdMcp(f))
```

## 8. Build & smoke test

```bash
make build
./jodoo-cli mcp serve <<'EOF'
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
EOF
```

Expected:
- One `initialize` response
- No response for the notification
- One `tools/list` response listing ~30 tools

Further tests in `06-testing.md`.

## 9. Phase-2 stubs to NOT forget

- `serveHTTP()` using the streamable-HTTP transport (POST + optional SSE stream)
- MCP `resources/*` exposing `jodoo://app/{id}/entry/{id}/schema` as a
  read-only resource → so Claude can preload a form's widget schema without
  a tool call
- MCP `prompts/*` with templates like "Summarise the last 100 records in {form}"
