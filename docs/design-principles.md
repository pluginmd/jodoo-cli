# Design Principles

This document captures the *why* behind the architectural decisions in
`jodoo-cli`. It is aimed at contributors who want to extend the CLI
without working against its grain.

## 1. The API shape determines the CLI shape

Jodoo's API is unusually regular:

- Every endpoint is `HTTP POST`.
- Every body is JSON.
- Every response wraps a payload in `{ "code": <int>, "msg": <string>, ...}`
  where `code == 0` means success.
- Auth is a single bearer token.

We lean into that regularity instead of generalizing. The result:

- The raw command is `jodoo-cli api <path>` — just a path, no `-X POST`,
  no method positional. One less flag the user has to learn.
  (See [`cmd/api/api.go`](../cmd/api/api.go).)
- The HTTP client is a single `APIClient.Do(ctx, Request)` function.
  Method is hard-coded to `POST`.
  (See [`internal/client/client.go`](../internal/client/client.go).)
- The dry-run preview prints `POST <url>` + body. No method line needed.
  (See [`shortcuts/common/dryrun.go`](../shortcuts/common/dryrun.go).)
- Error detection is a single envelope check. A `code: 0` response is
  success even on HTTP 400; a `code != 0` response is a failure even on
  HTTP 200 (Jodoo sometimes mixes these).

If a future Jodoo endpoint breaks these invariants (e.g. a `GET`, a
non-JSON response, a multipart body that isn't file upload), the
*design* stays the same but the client grows a second method — we
don't go back and refactor everything into a generic HTTP layer.

## 2. Shortcuts are declarative, not imperative

A shortcut is a `common.Shortcut` struct literal with four hooks
(`Validate`, `DryRun`, `Execute`, and optional metadata like `Risk`,
`Flags`, `Tips`). The runner owns the boring parts:

- Flag registration (string / int / bool / string_array) and cobra
  completion for `Enum` flags.
- `@path` / `-` stdin input resolution for flags that opt in via
  `Input: []string{File, Stdin}`.
- `--dry-run`, `--format`, `--jq`, and `--yes` wiring.
- Error envelope rendering and exit codes.

The code lives in [`shortcuts/common/runner.go`](../shortcuts/common/runner.go)
and [`shortcuts/common/types.go`](../shortcuts/common/types.go). The
pattern gives us:

- **Uniformity.** Every shortcut behaves the same way on the user-facing
  edges. You cannot accidentally ship a new command that forgets
  `--dry-run`.
- **Low contributor cost.** A new shortcut is ~30–60 lines of data +
  one `Execute` closure. It does not touch cobra, IO, or transport.
- **Agent-friendliness.** The skill doc at
  [`skills/working/jodoo-cli/SKILL.md`](../skills/working/jodoo-cli/SKILL.md)
  is a table of 30 rows because the commands *are* a table — not an ad
  hoc tree.

Trade-off: adding a feature to *every* shortcut is a single place
(the runner), so regression risk is concentrated. Tests + the
`+<name> --dry-run` smoke gate guard against it.

## 3. Escape hatch first

Every CLI that wraps a third-party API inherits the same decay: the
upstream ships a new endpoint, the wrapper falls behind, users get
stuck. `jodoo-cli api <path>` is the escape valve:

- Any JSON body, any path under the configured base URL.
- `--dry-run`, `--jq`, `--format`, and `--raw` all work.
- `--raw` keeps the envelope's `code` / `msg` fields so raw callers can
  tell the difference between a failed business call and a missing
  field.

The promise: you should never have to stop using the CLI because a
shortcut doesn't exist. You should, however, feel the pull to write
one — the shortcut is shorter to type and its `--help` is documented.

## 4. Fail closed, explain clearly

Every error becomes a structured envelope:

```json
{
  "ok": false,
  "error": {
    "type": "rate_limit",
    "code": 8303,
    "message": "Company/Team Request Limit Exceeded",
    "hint": "rate limit exceeded — back off and retry"
  }
}
```

Rules:

- **Types are stable** and safe for scripts to branch on:
  `auth | permission | rate_limit | validation | quota | member | form |
  data | workflow | department | api | network | config`.
  ([`internal/output/error.go`](../internal/output/error.go))
- **Exit codes map to categories**, not to specific errors:
  `2 config`, `3 auth`, `4 validation`, `5 api`, `6 network`.
- **Hints are remediation, not analysis.** `"run jodoo-cli config init"`
  not `"the config file doesn't exist"`.
- **Upstream codes flow through** untouched. If Jodoo returns `17026
  Duplicate transaction_id`, we say exactly that — not "something
  went wrong". The hint tells the user *what to do about it*.

We never swallow errors or synthesize a success out of partial data.

## 5. Safety is opt-out for reads, opt-in for writes

- Reads are free. `+app-list`, `+form-list`, `+data-get`, `+data-list`
  all run with no safety prompt.
- Regular writes still ship a `--dry-run` preview. Users are expected
  to use it when wiring a new automation; we don't force it.
- **High-risk writes** (`+data-batch-delete`,
  `+department-batch-import`, etc.) carry `Risk: "high-risk-write"` and
  refuse to run without `--yes`. This is not configurable — we never
  want a `JODOO_YES_ALWAYS=1` environment variable.

The runner enforces the gate centrally:
[`runner.go`](../shortcuts/common/runner.go) around the
`Risk == "high-risk-write"` check.

## 6. Credentials are boring

We treat the API key as a secret, but we don't build a key-management
system:

- **Environment first.** `JODOO_API_KEY` overrides config.json. This
  is what CI and Docker Compose use.
- **Config file second.** `~/.jodoo-cli/config.json` is mode `0600`,
  written atomically via `*.tmp → rename`. Directory is `0700`. Simple
  and portable.
- **OS keychain third.** Opt-in via `--use-keychain` during
  `config init`. macOS Keychain / libsecret / Windows Credential
  Manager via `zalando/go-keyring`. Falls back silently if the OS
  doesn't have one (e.g. headless CI on Linux).
  ([`internal/credential/credential.go`](../internal/credential/credential.go))

We deliberately do not support: `.env` auto-loading, gpg-encrypted
files, hashicorp vault integration. Anyone who needs those can wrap
the CLI.

## 7. One binary, small dependencies

`go.mod` has five direct dependencies. Each is load-bearing and
explicit:

- `spf13/cobra`+`pflag` — command framework.
- `itchyny/gojq` — pure-Go `jq` for `--jq` filtering (no shelling out).
- `google/uuid` — UUIDs for `transaction_id` auto-generation.
- `zalando/go-keyring` — OS keychain abstraction.
- `golang.org/x/term` — terminal capabilities (password prompts).

No DI framework, no code generation, no plugin loader. If we find
ourselves reaching for one, the feature is probably out of scope.

## 8. Output is for pipes first, humans second

The default format is `json` — the success envelope
`{ "ok": true, "data": ..., "meta": ... }`. Humans who find it noisy
reach for `--format pretty` or `--jq`. This ordering matters because:

- Scripts are the long-tail users. They should not have to fight the
  default.
- `--jq` removes the need for `--format table/csv` in 80 % of cases,
  and the jq expression lives in the script — readable forever.
- `--format table/csv/ndjson` exist for the 20 % where humans
  explicitly want flattened rows.
  ([`internal/output/format.go`](../internal/output/format.go))

When a shortcut has a natural human summary (e.g. `+widget-list` listing
fields), it provides a `prettyFn` that writes plain text. Otherwise
`pretty` falls back to indented JSON — never an error.

## 9. Tests verify contracts, not implementations

Where tests exist, they assert:

- The rendered envelope shape on success.
- The error envelope shape (type, code, hint presence) on each failure
  class.
- That every shortcut listed in `shortcuts.Shortcuts()` produces a valid
  `--dry-run` given its required flags.
- That `--jq` and `--format` don't mutate the underlying payload.

We do not assert HTTP headers, exact retry counts, or internal
sub-method calls. Refactors should be free; promises to users should
be fenced.
