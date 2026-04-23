# Code Standards

Short, opinionated. Deviations should have an explicit reason.

## Language & tooling

- **Go 1.23+.** `go.mod` pins `go 1.23.0`. No `//go:build` tags outside
  of obvious platform helpers.
- **Formatting**: `gofmt` + `goimports`. CI gate should run `go vet`
  and `go test -race -count=1 ./...`. Makefile target: `make test`.
- **`go mod tidy`** before every commit that touches dependencies.
- **No generated code.** If we ever need it (e.g. a typed client
  generated from OpenAPI), gate the generator behind `make gen` and
  keep the output checked in.

## Package shape

- **`cmd/` owns cobra.** No cobra imports under `internal/` or
  `shortcuts/`. Commands construct plumbing, then hand off.
- **`internal/` is unexported to the outside world.** This is a CLI,
  not a library. Anyone who wants a library builds their own on top of
  the Jodoo HTTP spec.
- **`shortcuts/common` is the framework.** `shortcuts/jodoo` is data.
  Never import `jodoo` from `common`, ever.
- **`internal/output` is the only package that writes to stdout or
  stderr.** Other packages return errors and structured data.

## Naming

- **Files**: snake_case only for multi-word filenames (`set_default.go`).
  Single-word filenames are fine (`runner.go`).
- **Exported types** use `CamelCase`; package-level vars / consts the
  same. Shortcut literals are `lowerCamelCase` (`appList`, `dataCreate`)
  because they are package-local.
- **Error constructors** live in `internal/output` and are named
  `Err<Category>` (`ErrAuth`, `ErrValidation`, `ErrAPI`,
  `ErrNetwork`, `ErrConfig`, `ErrWithHint`). Do not invent new ones
  without adding a matching exit code.
- **Shortcut commands** use `+<noun>-<verb>` kebab-case
  (`+data-list`, `+workflow-task-forward`). The `+` prefix is a
  deliberate visual cue that sets shortcuts apart from subcommand
  verbs (`auth status`, `profile use`).
- **Flag names** are kebab-case CLI-side (`--app-id`), snake_case
  body-side (`"app_id"`). The translation happens in `Execute`.

## Comments

- Default to none. If a comment adds nothing beyond the identifier, drop
  it.
- Write a comment when: (a) the *why* is not obvious from the code,
  (b) the function is exported and lives in `internal/` plumbing,
  (c) a subtle invariant would bite a future reader (e.g. "Jodoo
  returns HTTP 200 with `code != 0` on business errors — do not rely
  on the HTTP status alone").
- Package-level `// Package foo ...` docstrings on every package in
  `internal/` and `shortcuts/common/`.
- No banner comments (`// === Section ===`). Use `// ── Section ──` if
  the file really does have sub-sections worth visual separation (see
  `runner.go`), otherwise trust gofmt grouping.

## Error handling

- **Transport errors**: `output.ErrNetwork(format, args...)`. Wraps
  `dial`, `read`, `timeout`, etc. Exit code `6`.
- **API errors**: `output.ErrAPI(code, message, raw)`. Exit code `5`.
  The code is the Jodoo code, not the HTTP status. Always pass the raw
  response map when available so `--format json` carries the full
  server payload.
- **Validation errors (client-side)**: `output.ErrValidation(format,
  args...)`. Exit code `4`.
- **Config errors**: `output.ErrConfig` or `core.ConfigError` (the
  latter is auto-upgraded by `Factory.MustConfig`). Exit code `2`.
- **Auth errors**: `output.ErrAuth`. Exit code `3`.
- Never use `fmt.Errorf` at the CLI boundary. The boundary writes the
  envelope; `fmt.Errorf` prints a raw string and confuses scripts.
- Hints go in `error.jodooHint(code)`. If you learn a new pattern
  ("code X → remediation Y"), encode it once there.

## Logging

- No logging package. The CLI prints results to stdout and errors (as
  JSON) to stderr. That is the full output contract.
- Anything that needs a progress UI (`--paginate-all` with thousands
  of pages) writes to stderr with a line per page; stdout stays a
  clean, parseable stream.

## Testing

- Table tests are preferred. Keep the table small; one table per
  behavior.
- Test the *contract*, not the internals:
  - Envelope shape (`ok`, `data`, `meta`, `error.type`, `error.code`).
  - Exit codes for each error class.
  - Flag parsing (required flags, enum validation, `@path` / stdin
    resolution).
  - `--dry-run` output for every shortcut (smoke-level).
- Do not mock `net/http` with exotic frameworks. `httptest.NewServer`
  is enough.
- `go test -race` must pass. If a shared-mutable structure appears in
  the runner, fix it — do not add `t.Skip`.

## Security

- **Never log the API key.** Dry-run preview redacts it
  (`Bearer ab***yz`). If you introduce new debug output, keep it
  redacted.
- **File inputs are validated** via `internal/validate/path.go`. A
  path that escapes the working directory (`../../etc/passwd`) is
  rejected.
- **File-size cap** on `@path` inputs is 16 MB
  (`shortcuts/common/runner.go::maxInputSize`). Raise it only with a
  reason.
- **No shelling out.** If we ever need to call another binary, use
  `exec.Command` with a fixed argv and never interpolate user input.
- **No eval.** `--jq` runs inside `itchyny/gojq` — gojq is a safe
  interpreter; do not replace it with `jq`-shelling.

## Dependencies

Adding a dependency requires:

1. A pass at solving the problem with the standard library.
2. If the dep stays, a one-line rationale in the PR description
   ("needed because …"). We keep `go.mod` small on purpose.
3. A license check (MIT / BSD / Apache-2 only).

## Performance

- Allocation-hot paths (envelope rendering, pagination loops) use
  `make([]T, 0, hint)` where the hint is known.
- Do not precompute or cache anything beyond `Factory.config`. The CLI
  is a short-lived process; warm caches complicate reasoning.
- Timeouts: HTTP default is 30 s (`client.DefaultTimeout`), long
  operations use 5 min (`client.LongTimeout`). Shortcuts can override
  per-request via `client.Request{Timeout: ...}`.

## Docs

- Every user-facing change updates at least one of:
  - `README.md` (top-level surface).
  - `docs/jodoo-sdk-docs/` (API spec — keep in sync with what the
    shortcut actually calls).
  - `docs/adding-a-shortcut.md` (if you introduced a new pattern).
  - `skills/working/jodoo-cli/SKILL.md` (agent cheat-sheet).
- A change that moves a file or renames a package updates
  `docs/codebase-map.md`.

## Commits & PRs

- **Conventional commits**: `feat:`, `fix:`, `refactor:`, `docs:`,
  `chore:`, `test:`.
- Optional scope: `feat(shortcuts): add +role-member-list`,
  `fix(client): handle empty body with HTTP 204`.
- Keep commits small. A commit that touches `runner.go` should not
  also rename a shortcut.
- PRs link the relevant spec section under `docs/jodoo-sdk-docs/` so
  reviewers can diff against the source of truth.

## Versioning

- We follow SemVer. `internal/build.Version` is stamped by the Makefile
  via `git describe --tags --always --dirty`.
- Breaking changes to the CLI surface (renamed shortcut, renamed flag,
  changed exit code) are **major** bumps. We add deprecation warnings
  (to stderr only) for one minor cycle before removal.
- Additive changes (new shortcut, new optional flag, new error hint)
  are **minor**.
- Bug fixes and docs are **patch**.
