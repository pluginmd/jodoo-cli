# Project Overview

`jodoo-cli` is a focused, single-binary Go CLI for the
[Jodoo](https://api.jodoo.com) platform (氚数 / 简道云-style low-code forms
and workflow). It exposes roughly 30 curated shortcuts that cover the
surface area most teams actually touch — apps, forms, records, files,
workflow, contacts — and keeps a `jodoo-cli api` escape hatch for
endpoints that are not yet wrapped.

## Why this exists

Jodoo's HTTP API is regular: every endpoint is `POST`, every body is
JSON, and the response always follows a `{code, msg, ...payload}`
envelope. The UI and SDK teams already have full coverage, but we
repeatedly found ourselves reaching for one of three things:

- **Automation scripts** that need to list apps, pull records, or push
  data changes from CI or cron.
- **Shell-first exploration** (`curl | jq`) that is tedious to write
  against the bearer-token endpoints with strict filter/pagination
  semantics.
- **LLM agents** (Claude Code and friends) that need a typed,
  documentable surface so they stop inventing shapes from partial
  documentation.

`jodoo-cli` collapses all three use cases into one binary. The same
shortcut that a human runs interactively is what a script pipes into
`jq`, and what an AI skill (`skills/working/jodoo-cli`) describes to
the model.

## Scope

In scope:

- All **read** APIs documented under `docs/jodoo-sdk-docs/` (apps,
  forms, widgets, records, workflow instances/tasks/CC, comments,
  members, departments, roles).
- All **write** APIs with a first-class safety story (`--dry-run` by
  default, `--yes` for high-risk writes like `+data-batch-delete`,
  `+department-batch-import`).
- The two-step **file upload** flow (get token → upload →
  bind via `transaction_id`).
- **Auth / config / profile** management, including an optional OS
  keychain backend.
- **Doctor** connectivity check.
- **Output shaping**: `json`, `pretty`, `table`, `csv`, `ndjson`, and a
  gojq filter (`--jq` / `-q`).

Out of scope — by design:

- A full SDK. The CLI is the surface; library users should target the
  HTTP API directly (the specs in `docs/jodoo-sdk-docs/` are enough).
- Framework-level features like plugins, custom output templates, or
  structured TUI modes. We aim for Unix-pipe composability, not a new
  framework.
- Anything that persists user data beyond `~/.jodoo-cli/config.json`
  and the OS keychain. No telemetry. No autoupdate.

## Audience

1. **Operations & automation engineers** running data migrations,
   nightly record sync, or batch file uploads.
2. **Platform developers** who need a thin, dependable client while
   their product teams work on richer integrations.
3. **AI agents** (Claude Code skills under `skills/working/`) that
   benefit from typed flags, `--dry-run` previews, and stable
   machine-readable error envelopes.
4. **Curious users** exploring the Jodoo API without writing a single
   line of code.

## Success criteria

- A new user can go from `make install` to a green `jodoo-cli doctor`
  in under two minutes.
- Any endpoint under `docs/jodoo-sdk-docs/` can be reached either via
  a curated shortcut **or** via `jodoo-cli api <path>` — no exceptions.
- The CLI exits with a documented, structured JSON error envelope on
  every failure path, including transport-level errors.
- Adding a new shortcut requires one new `common.Shortcut` literal and
  zero changes to the root cobra wiring (see
  [`adding-a-shortcut.md`](./adding-a-shortcut.md)).

## Non-goals and trade-offs

- **No long-lived session state.** Profiles are flat. There is no
  refresh-token dance because Jodoo's token model does not need one.
- **No automatic retries on rate limits.** The error envelope surfaces
  `type: "rate_limit"` (codes 8303/8304) with a hint and exits — the
  caller decides the back-off policy.
- **No partial writes as a success.** Batch endpoints surface
  `success_count` / `success_ids` so scripts can inspect, but the exit
  code reflects the API envelope (`code != 0` is a hard failure).
- **No hidden merging of flags.** If both `--data` and a `--data-foo`
  shortcut flag overlap, the explicit flag wins and we document it;
  ambiguity is a bug.
