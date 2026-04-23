# Adding a Shortcut

This is the cookbook for wrapping a new Jodoo endpoint as a `jodoo-cli`
shortcut. Follow it end-to-end for anything beyond the simplest GETs.

The running example: we want to add the *Workflow Instance Reactivate*
API (`POST /v5/workflow/instance/activate`, parameters `instance_id`
and `flow_id`). It already exists in the codebase as
`workflowInstanceActivate`, so use it as a reference.

## 1. Confirm the spec

Read the API in [`docs/jodoo-sdk-docs/`](./jodoo-sdk-docs/):

- **Path** and **HTTP method** (always POST for Jodoo).
- **Request parameters** — name, type, required/default.
- **Response shape** — is it a list of objects, a single object, or
  just a status? This decides whether you need `OutFormat`'s pretty
  renderer.
- **Rate limit** — informational, but mention it in the description if
  it is unusual.
- **Error codes** specific to this endpoint — if there is a useful
  hint, add it to `internal/output/error.go` (`jodooHint`).

## 2. Pick the file

`shortcuts/jodoo/` is grouped by area:

| Area | File |
|---|---|
| Apps, forms | `app.go` |
| Records (data), widgets, filters | `data.go` |
| Files | `file.go` |
| Workflow | `workflow.go` |
| Contacts — members / departments / roles | `contact.go` |

If you can't find a natural home, open a new file and add it to
`shortcuts.go`. Keep one file per major area — do not sprawl.

## 3. Declare the shortcut

Shortcuts are package-level `var`s of type `common.Shortcut`. Use
helpers from `shortcuts/common/helpers.go` whenever possible so you
inherit descriptions, defaults, and required flags consistently.

```go
// Workflow Instance Reactivate — POST /v5/workflow/instance/activate
var workflowInstanceActivate = common.Shortcut{
    Service:     "jodoo",
    Command:     "+workflow-instance-activate",
    Description: "Reactivate a completed workflow instance",
    Risk:        "write",
    HasFormat:   true,
    Flags: []common.Flag{
        {Name: "instance-id", Desc: "instance (= data) ID", Required: true},
        {Name: "flow-id", Type: "int", Desc: "node ID to reactivate", Required: true},
    },
    DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
        return common.NewDryRunAPI("/v5/workflow/instance/activate").
            Set("instance_id", r.Str("instance-id")).
            Set("flow_id", r.Int("flow-id"))
    },
    Execute: func(_ context.Context, r *common.RuntimeContext) error {
        body := map[string]interface{}{
            "instance_id": r.Str("instance-id"),
            "flow_id":     r.Int("flow-id"),
        }
        data, err := r.CallAPI("/v5/workflow/instance/activate", body)
        if err != nil {
            return err
        }
        r.OutFormat(data, nil, nil)
        return nil
    },
}
```

Then wire it into the area slice:

```go
func workflowShortcuts() []common.Shortcut {
    return []common.Shortcut{
        /* ... */
        workflowInstanceActivate,
    }
}
```

No cobra changes. No register changes. The bundle is auto-mounted.

## 4. Flag conventions

- **Names are kebab-case** at the CLI (`--app-id`, `--transaction-id`).
  Body keys use the API's snake_case (`app_id`, `transaction_id`).
- **Type defaults to `"string"`.** Explicit types only for `bool`,
  `int`, `string_array`.
- **Required vs. default.** Prefer `Required: true` for identifiers
  (`--app-id`, `--entry-id`, `--data-id`, `--username`). Use `Default`
  for paging (`--limit`, `--skip`) and known-safe toggles.
- **Enums.** Declare `Enum: []string{...}` for server-side enums (e.g.
  `add-sign-type: [pre|post|parallel]`). The runner validates values
  and registers shell completion.
- **Reusable flags.** Use the helpers:
  - `common.AppIDFlag(required)`, `common.EntryIDFlag(required)`,
    `common.DataIDFlag(required)`.
  - `common.LimitFlag(default)`, `common.SkipFlag()`.
  - `common.UsernameFlag(required)`.
  - `common.DataJSONFlag(required)`, `common.DataListJSONFlag(required)`.
  - `common.FilterJSONFlag()`, `common.FieldsFlag()`.
  - `common.TransactionIDFlag(required, helpHint)`.
- **Input from file/stdin.** For flags that accept JSON blobs, opt in
  via `Input: []string{common.File, common.Stdin}`. The runner rewrites
  `@path` and `-` for you.

## 5. Risk tier

Set `Risk` when the call mutates data:

- `""` or `"read"` — pure read. No prompt.
- `"write"` — ordinary write. `--dry-run` is always available; no
  forced confirm.
- `"high-risk-write"` — irreversible or broad (`+data-batch-delete`,
  `+department-batch-import`). The runner refuses to run without
  `--yes`.

If you add a new tier, extend the runner in
`shortcuts/common/runner.go` and document it in
[`design-principles.md`](./design-principles.md).

## 6. `DryRun` is mandatory for writes

A `DryRun` hook prints the exact `POST <url>` + body the server will
receive. It must use the same flag reading logic as `Execute` so the
preview cannot diverge. The preview path redacts the bearer token
automatically.

```go
DryRun: func(_ context.Context, r *common.RuntimeContext) *common.DryRunAPI {
    return common.NewDryRunAPI("/v5/workflow/instance/activate").
        Set("instance_id", r.Str("instance-id")).
        Set("flow_id", r.Int("flow-id"))
},
```

Read-only shortcuts can omit `DryRun`, but we still recommend one — it
is cheap and helps users verify filter/body shapes without spending
rate-limit budget.

## 7. `Execute` patterns

### Simple call

```go
body := map[string]interface{}{ /* ... */ }
data, err := r.CallAPI("/v5/...", body)
if err != nil {
    return err
}
r.OutFormat(data, nil, nil)
return nil
```

### When the response *is* the envelope (no payload)

Some workflow actions return only `{"status":"success"}`. Use
`CallAPIRaw` to keep `code` and `msg`:

```go
data, err := r.CallAPIRaw("/v5/workflow/task/forward", body)
```

### When the response is a single object you want pretty-printed

Pass a closure as the third argument to `OutFormat`:

```go
r.OutFormat(data, meta, func(w io.Writer) {
    fmt.Fprintf(w, "instance %s → %s\n",
        data["instance_id"], data["status"])
})
```

### When pagination is needed

`/v5/app/entry/data/list` uses cursor pagination on `data_id`. For
read-all, prefer `r.PaginateAll`:

```go
items, err := r.PaginateAll(
    "/v5/app/entry/data/list", "data", pageSize,
    func(lastID string) map[string]interface{} {
        m := map[string]interface{}{
            "app_id":   r.Str("app-id"),
            "entry_id": r.Str("entry-id"),
            "limit":    pageSize,
        }
        if lastID != "" {
            m["data_id"] = lastID
        }
        if filterBody != nil {
            m["filter"] = filterBody
        }
        return m
    },
)
```

Always gate `--paginate-all` behind a flag and document the rate limit
in the description. (Jodoo's list endpoint is 5 req/s.)

### Validation

Put pre-flight checks into `Validate` rather than `Execute` — it runs
before `--dry-run`, so the preview can't be generated with bad input.

```go
Validate: func(_ context.Context, r *common.RuntimeContext) error {
    if r.Int("flow-id") <= 0 {
        return output.ErrValidation("--flow-id must be a positive integer")
    }
    return nil
},
```

Validation errors must use `output.ErrValidation` so the envelope gets
`type: "validation"` and exit code `4`.

## 8. Output

- **Default `--format json`** — `r.Out(data, meta)` or
  `r.OutFormat(data, meta, nil)`. Emits the
  `{ok, data, meta}` envelope.
- **`pretty`** — pass a closure to `OutFormat`'s third parameter that
  writes plain text. Fall through to JSON if you have nothing better
  (built-in).
- **`table` / `csv` / `ndjson`** — handled by `OutFormat` via
  `output.FormatValue` → `extractRows`. The extractor looks for known
  array keys (`apps`, `forms`, `widgets`, `data`, `list`, `users`,
  `departments`, `roles`, `tasks`, `members`, `logs`, `cc_list`,
  `approveCommentList`, `data_list`, `success_ids`, `results`). If
  your response uses a different key, either teach the extractor about
  it or shape the data before calling `OutFormat`.
- **`--jq`** — free. Users get `-q '.data.something[0]'` out of the box.

## 9. Help polish

- **Description**: start with a verb, no trailing period. Keep under 60
  chars. It shows up in `--help` and the agent skill.
- **Tips**: optional. Add `Tips: []string{"..."}` to surface follow-up
  commands or gotchas in `--help`. Common tips:
  - "pair with `+file-get-token` for image/attachment fields"
  - "rate limited to 5 req/s — combine with `--paginate-all` cautiously"
- Update [`skills/working/jodoo-cli/SKILL.md`](../skills/working/jodoo-cli/SKILL.md):
  add a row to the cheat-sheet table and, if the endpoint introduces a
  new pattern, a short narrative section.

## 10. Smoke test

```bash
go build -o /tmp/jodoo-cli .
# Structure — does `--help` render?
/tmp/jodoo-cli jodoo +workflow-instance-activate --help
# Preview — does --dry-run look right?
/tmp/jodoo-cli jodoo +workflow-instance-activate \
  --instance-id 123 --flow-id 2 --dry-run
# Actual (against a sandbox tenant):
JODOO_API_KEY=... /tmp/jodoo-cli jodoo +workflow-instance-activate \
  --instance-id 123 --flow-id 2
```

The dry-run must produce a valid `POST <url>` header and body that
matches the API spec verbatim. If it doesn't, the bug is in your
`DryRun` hook or your flag decoding.

## 11. Commit

- One shortcut per commit when possible.
- Conventional commit prefix: `feat(shortcuts): add
  +workflow-instance-activate`.
- If you touched the runner or a helper, use `feat(common): ...` and
  call out the impact radius in the body.
- If you updated the spec docs or the skill, include those in the same
  commit.

That's it. A good shortcut is small enough that a reviewer can verify
it by comparing your literal side-by-side with the spec file. If
your shortcut is growing custom helpers in `Execute`, consider
extracting them into `shortcuts/common/helpers.go` so the next wrapper
is shorter than yours.
