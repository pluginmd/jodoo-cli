---
name: jodoo-shared
description: Shared auth / config / scope / safety guidance for jodoo-cli. Loaded as a peer of jodoo-cli when the agent operates against the Jodoo platform.
when_to_use: Before running any `jodoo-cli` command for the first time, when the user asks "where is the API key stored", when an `auth` / `config` / `permission` error appears, or when migrating profiles between machines.
---

# Jodoo Shared — auth, config, safety

This skill is the small, stable foundation that every other Jodoo skill
inherits from. It explains where credentials live, what the env vars do,
and which Jodoo error codes mean "you cannot proceed without admin help".

## Auth model

Jodoo authentication is a single Bearer token — the **API Key** generated
in Open Platform. There is no OAuth flow, no refresh dance, no scope
grant flow. One key per company can be created up to 500 times; each key
is scoped to a configurable subset of apps.

```http
Authorization: Bearer YOUR_APIKEY
Content-Type:  application/json
```

`jodoo-cli` injects both headers on every request. You should never need
to construct them manually.

## Where the key is stored

Resolution order (first non-empty wins):

1. **`JODOO_API_KEY` env var** — best for CI / containers. Always
   overrides the on-disk config.
2. **`~/.jodoo-cli/config.json`** (mode 0600) — the default for
   `jodoo-cli config init`. Profile entries look like:

   ```json
   {
     "default": "default",
     "profiles": {
       "default": { "profile": "default", "api_key": "sk-...", "base_url": "https://api.jodoo.com/api" }
     }
   }
   ```

3. **OS keychain** — opt-in via `--use-keychain` on `config init` /
   `auth set`. Service name `jodoo-cli`, account `apikey:<profile>`.
   Useful when corporate policy forbids plaintext keys on disk.

`base_url` is configurable so you can point at a private region or staging
Jodoo cluster (`JODOO_BASE_URL` overrides at runtime).

## Profiles in plain English

A profile is a (name → API key + base URL + notes) triple. Each shell
command resolves the profile in this order:

1. `--profile NAME` flag
2. `JODOO_PROFILE=NAME` env
3. `default` field in `config.json`
4. Literal `"default"`

So:

```bash
jodoo-cli profile use companyA              # change the file default
JODOO_PROFILE=companyB jodoo-cli doctor     # one-shot env override
jodoo-cli --profile companyC jodoo +app-list # one-shot flag override
```

## Safety: high-risk-write commands

The following shortcuts are flagged `high-risk-write` and refuse to run
without `--yes`:

- `+data-delete`, `+data-batch-delete`
- `+member-delete`
- `+department-delete`, `+department-batch-import`
- `+workflow-instance-close`

Always preview them with `--dry-run` first; only add `--yes` once you've
read the printed body.

## Rate limits

Per-endpoint limits live in the response — Jodoo returns code `8303`
(team-wide) or `8304` (per-API). The CLI surfaces them under `error.type
= "rate_limit"`. Sensible defaults to remember:

- `+data-list` is the slowest: **5 req/s**
- batch create / update: **10 req/s**
- single-record CRUD: **20 req/s**
- query endpoints (apps, forms, members, workflow get): **30 req/s**

When `--paginate-all` walks a large form, expect to hit `5 req/s` —
either narrow the filter or add a sleep between calls.

## Common error codes

| Code | Type | Meaning | Remediation |
|---|---|---|---|
| 8301 / 17018 | auth | invalid API key | `jodoo-cli auth set` or fix `JODOO_API_KEY` |
| 8302 | permission | key not authorized for that app | re-scope on Open Platform |
| 8303 / 8304 | rate_limit | global / per-API throttle | back off, retry with jitter |
| 17017 / 17032 / 17034 | validation | bad params / unknown field type | check `+widget-list` for the actual `_widget_*` IDs |
| 17025 | validation | bad transaction_id | must be a UUID |
| 17026 | validation | duplicate transaction_id | use a fresh UUID; idempotency only protects against in-flight dupes for the same body |
| 17023 / 17024 | quota | batch too large | split into ≤100/req chunks |
| 4815 | validation | malformed filter | use the documented `{rel,cond:[{field,type,method,value}]}` shape |
| 7212 / 7216-7219 | quota | monthly data / attachment quota | upgrade plan or contact business owner |

## Field naming convention

Form fields are immutable IDs prefixed `_widget_` (e.g.
`_widget_1529400746031`). Aliases configured in **Extension → Webhook →
Set Field Alias** replace those IDs everywhere — webhook payloads AND
API requests / responses. If an alias exists, scripts must use the alias.

## Time formats accepted everywhere

- ISO 8601: `2018-11-09T10:00:00Z` or `2018-11-09T10:00:00`
- Milliseconds since epoch: `1639106951523` (must be ms, not s)
- RFC 3339 / `yyyy-MM-dd HH:mm:ss` or `yyyy-MM-dd`

When in doubt: use ISO 8601 with `Z`.

## What this skill does NOT cover

- The contents of any individual shortcut → see `jodoo-cli` skill.
- Webhook subscription wiring (Jodoo configures webhooks via UI, not API).
- SSO / SAML / SCIM (those are admin-console only at the moment).
