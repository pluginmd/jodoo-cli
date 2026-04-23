# Codebase Map

A depth-first tour of the repository. Line counts are approximate and
drift with edits; the important part is the layering.

```
JodooCLI/
├── main.go                      # 15 loc — only calls cmd.Execute()
├── go.mod / go.sum              # 5 direct deps (cobra, pflag, gojq, uuid, keyring)
├── Makefile                     # build / install / test / tidy targets
├── build.sh                     # one-shot portable build
│
├── cmd/                         # cobra entry points
│   ├── root.go                  # wiring + tips + root error handler
│   ├── bootstrap.go             # pre-cobra --profile scan
│   ├── global_flags.go          # --profile registration
│   ├── api/api.go               # raw POST <path> escape hatch
│   ├── auth/{auth,set,status,clear}.go
│   ├── config/{config,init,show,remove,set_default}.go
│   ├── doctor/doctor.go         # version + config + key + connectivity
│   ├── mcp/                     # Model Context Protocol layer (Claude Desktop, …)
│   │   ├── mcp.go               # cobra `mcp serve` + flags
│   │   ├── server.go            # stdio JSON-RPC 2.0 loop
│   │   ├── http.go              # streamable-HTTP transport (phase 2)
│   │   ├── tools.go             # registry + tools/list + tools/call + risk gate
│   │   ├── schema.go            # common.Flag[] → JSON Schema
│   │   ├── dispatch.go          # fork-self argv builder + subprocess exec
│   │   └── log.go               # file/stderr logger (never logs arguments)
│   └── profile/profile.go       # list / use / rename / remove
│
├── internal/
│   ├── build/build.go           # Version, Date — set via -ldflags
│   ├── client/client.go         # APIClient + file upload
│   ├── cmdutil/
│   │   ├── factory.go           # lazy config + client, profile cache
│   │   ├── iostreams.go         # stdio abstraction
│   │   └── tips.go              # per-command help tips
│   ├── core/config.go           # CliConfig, ConfigFile, env vars
│   ├── credential/credential.go # OS keychain glue
│   ├── output/
│   │   ├── format.go            # envelope, table, csv, ndjson, column priority
│   │   ├── error.go             # ExitError, jodooType/jodooHint maps
│   │   └── jq.go                # gojq bridge + flag validation
│   └── validate/path.go         # SafeInputPath (reject traversal)
│
├── shortcuts/
│   ├── register.go              # mounts every bundle on root
│   ├── common/
│   │   ├── types.go             # Shortcut, Flag, File/Stdin constants
│   │   ├── runner.go            # the heart of the framework (~480 loc)
│   │   ├── dryrun.go            # DryRunAPI builder (POST <url> + body)
│   │   └── helpers.go           # shared flag factories + transaction_id UUID
│   └── jodoo/
│       ├── shortcuts.go         # concat of the five area lists
│       ├── app.go               # +app-list, +form-list
│       ├── data.go              # record CRUD + batch + filter/pagination
│       ├── file.go              # +file-get-token, +file-upload
│       ├── workflow.go          # instance / task / CC / comments
│       └── contact.go           # member / department / role
│
├── skills/working/              # AI agent skills (Claude Code)
│   ├── jodoo-cli/SKILL.md       # command reference for agents
│   └── jodoo-shared/SKILL.md    # conventions shared across future skills
│
└── docs/
    ├── jodoo-sdk-docs/          # authoritative API specs (8 files)
    ├── MCP/                     # MCP integration guide (9 files)
    ├── project-overview.md      # this series
    ├── design-principles.md
    ├── system-architecture.md
    ├── codebase-map.md          ← you are here
    ├── adding-a-shortcut.md
    └── code-standards.md
```

## Entry point

[`main.go`](../main.go) is a one-liner that calls `cmd.Execute()` and
uses the returned exit code. Do not add logic here — it makes test
harnesses awkward.

## cmd/ — command layer

### `cmd/root.go`

Builds the root cobra command, registers the five top-level commands
(`config`, `auth`, `profile`, `doctor`, `api`), and then delegates to
`shortcuts.RegisterShortcuts` for the `jodoo` bucket. Also owns two
cross-cutting behaviors:

- `installTipsHelpFunc` appends a "Tips:" block to `--help` when a
  command registered any via `cmdutil.SetTips`.
- `handleRootError` converts every error to an `*output.ExitError`
  (catching `*core.ConfigError` on the way) and renders the standard
  JSON envelope on stderr.

### `cmd/bootstrap.go`

Runs before cobra. It manually scans `os.Args` for `--profile` (both
`--profile foo` and `--profile=foo` forms) so the `Factory` knows the
active profile *before* any cobra hook fires. All other flag parsing
stays with cobra.

### `cmd/api/api.go`

The raw-POST escape hatch. Three things worth knowing:

- `--data` supports `@path`, `-` (stdin), and `@@...` for a literal
  leading `@`. The resolver lives at the top of the file.
- `--raw` toggles between returning only the payload (default) and the
  full `{code, msg, ...}` envelope.
- On error, the underlying `*output.ExitError` is marked
  `Raw: true` so the root handler leaves its `Detail.Detail` intact
  (raw callers want the raw upstream shape).

### `cmd/mcp/`

The Model Context Protocol layer. Exposes every shortcut as an MCP tool so
Claude Desktop / Claude Code / Cursor / Zed can drive Jodoo with the same
curated surface a human uses on the terminal. Two transports: stdio
(`jodoo-cli mcp serve`) and streamable HTTP (`jodoo-cli mcp serve --http :port`).
Tool calls fork `jodoo-cli` as a subprocess — identical behaviour to the
terminal, zero flag-storage races. Full guide in [`docs/MCP/`](./MCP/).

### `cmd/auth/`, `cmd/config/`, `cmd/profile/`

Thin wrappers around `internal/core` + `internal/credential`. They do
I/O (`config.json` or keychain) and never call the Jodoo API. Perfect
targets for adding more profile-management UX.

### `cmd/doctor/doctor.go`

Five checks in fixed order: version, config file, profile, API key,
connectivity. Supports `--json` for CI. Connectivity hits
`/v5/app/list` with `limit=1` — cheapest read Jodoo exposes.

## internal/ — plumbing

### `internal/core/config.go`

- `CliConfig` — one profile (api_key, base_url, notes).
- `ConfigFile` — `{default: string, profiles: map[name]*CliConfig}`.
- Env vars: `JODOO_API_KEY`, `JODOO_BASE_URL`, `JODOO_PROFILE`,
  `JODOO_CLI_HOME`.
- `LoadResolved(profile)` is the one-stop resolver. Returns a
  `*ConfigError` with `Code: 2` when no key is reachable.
- On-disk layout: `~/.jodoo-cli/config.json` mode `0600`, parent
  directory mode `0700`. Writes are atomic (`.tmp → rename`).

### `internal/credential/credential.go`

Wraps `zalando/go-keyring`. Service name is hardcoded to `jodoo-cli`,
account is `apikey:<profile>`. `Resolve(profile, cfg)` is the fallback
used by `Factory.Config` when config.json doesn't have a key.

### `internal/client/client.go`

The HTTP client. Two request paths:

- `Do(ctx, Request{Path, Body, Headers, Timeout})` — every normal
  endpoint. Returns `*RawResponse` with parsed `Code`, `Msg`, full
  `Data` map, and the raw body. On `Code != 0` **or** `Status >= 400`,
  returns an `*output.ExitError` from `ErrAPI`.
- `UploadFile(ctx, FileUploadRequest{URL, Token, FilePath|FileBytes,
  FileName, Timeout})` — multipart upload to the per-token URL returned
  by `+file-get-token`. 5-minute default timeout.

`BuildDryRun` returns a redacted preview (`Bearer ab***yz`) used by
`jodoo-cli api --dry-run` and every shortcut's dry-run hook.

### `internal/output/`

Three files:

- `format.go` — `WriteEnvelope`, `PrintJson`, `FormatValue` (table /
  csv / ndjson), and `extractRows` that finds an array-of-objects in
  Jodoo payloads by preferring known keys (`apps`, `forms`, `widgets`,
  `data`, `list`, `users`, `departments`, `roles`, `tasks`, `members`,
  `logs`, `cc_list`, …).
- `error.go` — `ExitError`, `ErrAuth` / `ErrConfig` / `ErrValidation` /
  `ErrAPI` / `ErrNetwork`, and the `jodooType` / `jodooHint` maps.
- `jq.go` — `JqFilter` + `ValidateJqFlags` (rejects `--jq` combined
  with `table` / `csv` / `ndjson` where it would silently misbehave).

### `internal/cmdutil/factory.go`

`Factory` is the dependency container. It:

- Lazily loads config on the first `Config()` call.
- Falls back to the keychain transparently when config.json has no key.
- Caches across a single CLI invocation.
- Offers `MustConfig()` which upgrades a `*core.ConfigError` into an
  `*output.ExitError` so the root handler can render it.

### `internal/validate/path.go`

`SafeInputPath` is used by `@path` resolution (in both `cmd/api` and
`shortcuts/common/runner`). It rejects traversal attempts and enforces
that the path is a regular file under the user's scope.

## shortcuts/ — the framework

### `shortcuts/common/types.go`

Only two types:

- `Flag{Name, Type, Default, Desc, Hidden, Required, Enum, Input}` —
  fully declarative.
- `Shortcut{Service, Command, Description, Risk, Flags, HasFormat,
  Tips, DryRun, Validate, Execute}` — the whole shortcut contract.

### `shortcuts/common/runner.go`

Central piece. Responsibilities in order:

1. `Mount` wraps a `Shortcut` into a cobra subcommand.
2. `registerShortcutFlags` turns the declarative `Flag`s into cobra
   flags, including enum completion and `@path`/`-` hints in help text.
3. `runShortcut` is the request lifecycle: load config → set Format /
   Jq → enum validation → input resolution → `Shortcut.Validate` →
   `--dry-run` or `Risk` gate → `Shortcut.Execute`.
4. `RuntimeContext` is what every hook sees. Flag accessors
   (`Str/Bool/Int/StrArray`), API helpers (`CallAPI`, `CallAPIRaw`,
   `PaginateAll`), output helpers (`Out`, `OutFormat`).
5. `ParseJSONObject` / `ParseJSONArray` / `ParseIntBounded` —
   utilities used by shortcut `Execute` functions.

### `shortcuts/common/dryrun.go`

`DryRunAPI` is the preview returned by `Shortcut.DryRun`. It is a
`{url, body}` pair with a fluent builder (`Set`, `SetIf`, `BodyJSON`).
Method is always POST — not printed. `Format()` renders
`POST <url>\n<body JSON>\n`.

### `shortcuts/common/helpers.go`

Flag factories (`AppIDFlag`, `EntryIDFlag`, `DataIDFlag`, `LimitFlag`,
`SkipFlag`, `UsernameFlag`, `DataJSONFlag`, `DataListJSONFlag`,
`FilterJSONFlag`, `FieldsFlag`, `TransactionIDFlag`). Reuse these.
Also hosts `EnsureTransactionID(r)` which returns the user-supplied
`--transaction-id` or generates a fresh UUID v4 when empty.

### `shortcuts/jodoo/`

The 30+ shortcut definitions, grouped by API area. Naming follows the
spec in `docs/jodoo-sdk-docs/`:

| File | Shortcuts |
|---|---|
| `app.go` | `+app-list`, `+form-list` |
| `data.go` | `+widget-list`, `+data-get`, `+data-list`, `+data-create`, `+data-batch-create`, `+data-update`, `+data-batch-update`, `+data-delete`, `+data-batch-delete` |
| `file.go` | `+file-get-token`, `+file-upload` |
| `workflow.go` | `+workflow-instance-get`, `+workflow-instance-logs`, `+workflow-instance-activate`, `+workflow-instance-close`, `+workflow-task-list`, `+workflow-task-forward`, `+workflow-task-reject`, `+workflow-task-back`, `+workflow-task-transfer`, `+workflow-task-revoke`, `+workflow-task-add-sign`, `+workflow-cc-list`, `+approval-comments` |
| `contact.go` | `+member-list`, `+member-get`, `+member-create`, `+member-update`, `+member-delete`, `+member-batch-import`, `+department-list`, `+department-create`, `+department-update`, `+department-delete`, `+department-batch-import`, `+role-list`, `+role-create`, `+role-member-list` |

`shortcuts.go` assembles all of them in display order.
`shortcuts/register.go` sorts them alphabetically before mounting so
`--help` output is deterministic.

## docs/

- `jodoo-sdk-docs/` — authoritative API specs (previously monolithic
  `research.md`). Source of truth for shortcut implementations.
- `MCP/` — MCP (Model Context Protocol) integration guide: overview,
  architecture, implementation walk-through, tool catalog, Claude Desktop
  setup, security, testing, roadmap. Read alongside `cmd/mcp/`.
- `project-overview.md` — scope / goals / audience.
- `design-principles.md` — the *why*.
- `system-architecture.md` — layers, lifecycle, packages.
- `codebase-map.md` — this file.
- `adding-a-shortcut.md` — cookbook.
- `code-standards.md` — conventions.

## Where to start when…

- **You want to add an endpoint.** Read
  [`adding-a-shortcut.md`](./adding-a-shortcut.md). Start from the
  closest existing shortcut in `shortcuts/jodoo/*.go`, copy it, adjust.
- **You want to change error messages or hints.**
  [`internal/output/error.go`](../internal/output/error.go). Keep the
  `type` field stable.
- **You want to support a new output format.**
  [`internal/output/format.go`](../internal/output/format.go).
  Register it in `ParseFormat`, implement a writer, wire it into
  `FormatValue`.
- **You want to ship a new top-level command.** Add a sibling package
  under `cmd/` and register it in [`cmd/root.go`](../cmd/root.go).
- **You want to extend the framework (e.g. a new `Risk` tier).**
  [`shortcuts/common/runner.go`](../shortcuts/common/runner.go). Keep
  opt-in — do not break any existing shortcut literal.
