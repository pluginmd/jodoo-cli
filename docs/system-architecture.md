# System Architecture

This is a tour of the binary from first byte to last, with enough
pointers into the source tree that you can answer "where does *X*
live?" in one hop.

## Top-level view

```
              ┌─────────────────────────────────┐
              │ main.go → cmd.Execute()         │
              └─────────────────────────────────┘
                             │
           bootstrap (parse --profile early, build Factory)
                             │
                 ┌──────────────────────┐
                 │   cobra root command │
                 └──────────────────────┘
                             │
  ┌───────────┬───────────┬─────┴───────┬──────────────┬──────────────┬──────┐
  ▼           ▼           ▼             ▼              ▼              ▼      ▼
config      auth       profile       doctor            api           mcp   (shortcuts)
(CRUD on  (set/show/  (list/use/   (connectivity     (raw          (MCP
 ~/.jodoo  clear)      rename/      probe + key       POST          server,
 -cli/                 remove)      visibility)       to any        stdio
 config.                                              path)         + HTTP)
 json)
                             │
                             ▼
                ┌───────────────────────────┐
                │   shortcuts.Register      │
                │   (jodooProvider)         │
                └───────────────────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        jodoo +app-*   jodoo +data-*   jodoo +workflow-*
        jodoo +file-*  jodoo +member-* ... (~30 verbs)

                             │
                             ▼
                ┌───────────────────────────┐
                │  shortcuts/common runner  │  ← shared: flags, dry-run,
                │                           │    enum/input/jq/format,
                │                           │    risk gate, error render
                └───────────────────────────┘
                             │
                             ▼
                ┌───────────────────────────┐
                │   internal/client         │  Bearer-auth POST + JSON
                │   (APIClient.Do)          │  envelope parse
                └───────────────────────────┘
                             │
                             ▼
                  https://api.jodoo.com/api
```

## Packages by responsibility

| Package | Role | Key files |
|---|---|---|
| `cmd/` | Cobra entry points for top-level commands. | [`root.go`](../cmd/root.go), [`bootstrap.go`](../cmd/bootstrap.go), [`global_flags.go`](../cmd/global_flags.go) |
| `cmd/api/` | Raw `POST <path>` escape hatch. | [`api.go`](../cmd/api/api.go) |
| `cmd/auth/` | `auth set` / `status` / `clear`. Keychain vs. file logic lives here. | [`set.go`](../cmd/auth/set.go), [`status.go`](../cmd/auth/status.go), [`clear.go`](../cmd/auth/clear.go) |
| `cmd/config/` | `config init / show / remove / set-default`. | [`init.go`](../cmd/config/init.go), [`show.go`](../cmd/config/show.go) |
| `cmd/doctor/` | 5-step connectivity probe with `--json` output. | [`doctor.go`](../cmd/doctor/doctor.go) |
| `cmd/profile/` | Multi-tenant profile management. | [`profile.go`](../cmd/profile/profile.go) |
| `cmd/mcp/` | MCP server layer (stdio + HTTP) — exposes every shortcut as a Model Context Protocol tool. | [`mcp.go`](../cmd/mcp/mcp.go), [`server.go`](../cmd/mcp/server.go), [`http.go`](../cmd/mcp/http.go), [`tools.go`](../cmd/mcp/tools.go), [`schema.go`](../cmd/mcp/schema.go), [`dispatch.go`](../cmd/mcp/dispatch.go) |
| `internal/core/` | On-disk `ConfigFile` type, env vars, profile resolution. | [`config.go`](../internal/core/config.go) |
| `internal/credential/` | OS keychain glue (`zalando/go-keyring`). | [`credential.go`](../internal/credential/credential.go) |
| `internal/client/` | HTTP client, envelope decode, dry-run, file upload. | [`client.go`](../internal/client/client.go) |
| `internal/cmdutil/` | `Factory` (lazy config+client), IO streams, tips. | [`factory.go`](../internal/cmdutil/factory.go), [`tips.go`](../internal/cmdutil/tips.go) |
| `internal/output/` | Envelope, format (table/csv/ndjson/pretty), `--jq`, structured errors. | [`format.go`](../internal/output/format.go), [`error.go`](../internal/output/error.go), [`jq.go`](../internal/output/jq.go) |
| `internal/validate/` | Path safety (reject `..`, absolute escapes). | [`path.go`](../internal/validate/path.go) |
| `internal/build/` | `Version` / `Date` stamped via `-ldflags`. | [`build.go`](../internal/build/build.go) |
| `shortcuts/common/` | Declarative shortcut framework. | [`types.go`](../shortcuts/common/types.go), [`runner.go`](../shortcuts/common/runner.go), [`dryrun.go`](../shortcuts/common/dryrun.go), [`helpers.go`](../shortcuts/common/helpers.go) |
| `shortcuts/jodoo/` | ~30 shortcut declarations, grouped by API area. | [`app.go`](../shortcuts/jodoo/app.go), [`data.go`](../shortcuts/jodoo/data.go), [`file.go`](../shortcuts/jodoo/file.go), [`workflow.go`](../shortcuts/jodoo/workflow.go), [`contact.go`](../shortcuts/jodoo/contact.go) |
| `shortcuts/register.go` | Mounts every shortcut alphabetically under its service bucket. | [`register.go`](../shortcuts/register.go) |
| `skills/working/` | AI agent skills (Claude Code). Ship next to the binary. | [`jodoo-cli/SKILL.md`](../skills/working/jodoo-cli/SKILL.md), [`jodoo-shared/SKILL.md`](../skills/working/jodoo-shared/SKILL.md) |

## Request lifecycle

### Startup

1. `main.go` calls `cmd.Execute()`.
2. `BootstrapInvocationContext` manually scans `os.Args` for
   `--profile`. We do this *before* cobra runs so that the `Factory`
   can be constructed with the right profile without waiting for
   cobra's parser.
3. `cmdutil.NewDefault(inv)` wraps stdio + the profile override into a
   `*Factory`. Config loading is lazy — we defer it until some command
   needs a key.
4. `root.go` registers the top-level commands and then calls
   `shortcuts.RegisterShortcuts(root, f)` which mounts the `jodoo`
   bucket and every `+<verb>` under it (sorted alphabetically for
   stable `--help` output).

### Running a shortcut

(Running `jodoo-cli jodoo +data-list --app-id A --entry-id E`.)

1. Cobra parses flags; the `RunE` closure for the mounted command calls
   `common.runShortcut(cmd, factory, &shortcut)`.
2. `Factory.MustConfig()` loads the resolved profile on first call
   (env → file → keychain escalation), caches it, and returns an
   `*output.ExitError` with `Code: 2` if no key is set.
3. The runner reads `--format` / `--jq`, validates enum flags, expands
   `@path` / `-` for opt-in input sources, then runs:
   - `Shortcut.Validate` (optional pre-flight checks).
   - If `--dry-run`: `Shortcut.DryRun` returns a `*common.DryRunAPI`
     which is printed as `POST <url>` + body (or as JSON), and we exit.
   - If `Risk == "high-risk-write"` and `--yes` is missing: we exit
     with a validation error.
   - Otherwise `Shortcut.Execute` runs. Inside, the shortcut calls
     `r.CallAPI(path, body)`.
4. `RuntimeContext.CallAPI` builds a `client.Request`, hands it to
   `APIClient.Do`, and returns `resp.PayloadOnly()` — the data map with
   the `code` and `msg` envelope keys stripped.
5. `Shortcut.Execute` invokes `r.OutFormat(data, meta, prettyFn)` which
   dispatches to JSON / pretty / table / csv / ndjson and applies `--jq`
   if supplied.
6. On any error, the error propagates as `*output.ExitError`.
   `root.handleRootError` renders it as the standard error envelope on
   stderr and returns the exit code.

### Running the raw API command

(Running `jodoo-cli api /v5/corp/member/get --data '{"username":"alice"}'`.)

1. `loadBodyFlag` resolves the `--data` flag, supporting `@path`,
   `-` (stdin), and `@@...` (literal leading `@`).
2. `client.New(cfg).Do` sends the request and parses the envelope.
3. If the request failed at the transport level, we mark the returned
   `*output.ExitError` as `Raw: true` so the root handler preserves the
   upstream detail (type, code, message, hint).
4. On success, we write either `resp.PayloadOnly()` (default) or
   `resp.Data` (when `--raw` is set) through the same format/jq path
   shortcuts use.

### Running an MCP tool call

(An MCP client calls `tools/call` with `name: "jodoo.data-list"` and
`arguments: {"app_id":"A","entry_id":"E","limit":50}`.)

1. The transport (stdio reader or HTTP `POST /mcp`) parses the JSON-RPC
   frame and hands the request to the dispatcher.
2. The registry (built once at startup from `shortcuts.Providers()`) looks
   up the tool by name, applying `--read-only` and `--allow`/`--deny`
   filters. Unknown name → JSON-RPC `-32602`.
3. For `Risk == "high-risk-write"` shortcuts, the dispatcher checks
   `arguments.confirm == true` before doing anything else — fails fast with
   `isError: true` and no subprocess fork.
4. `dispatch.buildArgv` translates the arguments into the equivalent CLI
   argv (snake_case → kebab-case, array → repeated flags, object → JSON
   string, `dry_run` → `--dry-run`, `confirm` → `--yes`).
5. `exec.CommandContext(self, argv…)` forks a child `jodoo-cli` — identical
   execution path to the terminal (auth, validation, envelope, risk gate).
6. Child stdout becomes the tool result's `text` content; stderr + non-zero
   exit flips `isError: true`. The protocol-level response goes back on
   stdio (newline-delimited) or HTTP (one JSON body, or an SSE stream for
   progress notifications).

See [`docs/MCP/01-architecture.md`](./MCP/01-architecture.md) for the full
picture.

### File upload (two-step)

Jodoo's file upload does not fit the envelope pattern, so we expose it
as two shortcuts:

1. `+file-get-token` → `POST /v5/app/entry/file/get_upload_token`
   returns up to 100 `{url, token}` pairs scoped to a `transaction_id`.
2. `+file-upload` → `POST <url>` with a multipart body that carries
   `token` and `file`. `APIClient.UploadFile` handles the multipart
   encoding, targets the per-token URL (outside the configured base
   URL), and still parses the `{code, msg, key}` envelope.
3. The caller then sends `+data-create` / `+data-update` with the
   *same* `transaction_id` so Jodoo binds the uploaded file to the
   record.

If a user forgets the `transaction_id`, Jodoo returns code `17025`
and our error handler points at the exact remediation.

## Config resolution

The resolved profile is the output of a chain of overrides
(implementation in [`core.LoadResolved`](../internal/core/config.go)
and [`cmdutil.Factory.Config`](../internal/cmdutil/factory.go)):

```
  profile name ─── --profile flag   > JODOO_PROFILE env  > config default  > "default"
  api key      ─── JODOO_API_KEY    > config.json        > OS keychain (fallback)
  base url     ─── JODOO_BASE_URL   > config.json        > "https://api.jodoo.com/api"
```

The `--profile` flag is parsed *before* cobra runs so that every other
command sees the right profile (`bootstrap.go`). All other resolution
happens inside `Factory.Config()` and is cached per-process.

## Error envelope

Every failure path funnels through `*output.ExitError`. On exit:

- The root command writes the envelope to **stderr** and returns the
  exit code.
- Success payloads go to **stdout**.

Scripts should therefore branch on `$?` (exit code) for coarse
decisions and parse stderr JSON (`jq '.error.type'`) for fine-grained
handling. Combining the two is how our own agent skill reacts to
rate limits and retryable auth errors.

## AI agent surface

`skills/working/` ships alongside the binary. Two skills today:

- `jodoo-cli` — full command reference + when-to-use heuristics.
- `jodoo-shared` — shared conventions (profiles, env vars, output
  shapes) reused by future skills (e.g. a "Jodoo migration" skill).

The CLI is designed so that the skill Markdown is essentially the
`--help` of the binary plus usage narrative — no translation layer.
When the CLI gets a new shortcut, the skill table grows by one row.
