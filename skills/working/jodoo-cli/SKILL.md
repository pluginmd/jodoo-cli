---
name: jodoo-cli
description: Drive the Jodoo (api.jodoo.com) platform from the command line — apps, forms, records, files, workflows, contacts.
when_to_use: User mentions Jodoo, jodoo-cli, "竹间智能", `jodoo` command, Jodoo apps/forms/records/widgets, Jodoo workflow tasks/instances, Jodoo members/departments/roles, or any HTTP path under `/v5/app/*`, `/v5/workflow/*`, `/v5/corp/*`.
---

# Jodoo CLI — `jodoo-cli`

A Go CLI that wraps every Jodoo open API. All endpoints are HTTPS POST to
`https://api.jodoo.com/api`, authenticated by a single Bearer token (the
"API key" generated in **Open Platform → API Key**).

## When to reach for this skill

- The user says "list Jodoo apps / forms / records"
- The user wants to upload a file to a Jodoo form
- The user wants to manage Jodoo members / departments / roles
- The user wants to operate Jodoo workflow tasks (approve / reject / transfer)
- The user wants to call a Jodoo endpoint that is not yet wrapped (use `jodoo-cli api`)
- A `lark-*` skill matched but the URL is `api.jodoo.com` — wrong system, swap to this skill

## First-run checklist

Before any API call:

```bash
# 1. Configure (interactive prompts; key stored in ~/.jodoo-cli/config.json)
jodoo-cli config init                           # → uses default profile
jodoo-cli config init --profile prod --use-keychain   # → store in OS keychain

# 2. Sanity check
jodoo-cli doctor
```

If `JODOO_API_KEY` is exported in env, no setup is needed — it overrides
config.json. CI / automation should prefer that.

## Command map (cheat sheet)

| Goal | Shortcut |
|---|---|
| List apps reachable by the key | `jodoo +app-list` |
| List forms in an app | `jodoo +form-list --app-id <a>` |
| Inspect form fields | `jodoo +widget-list --app-id <a> --entry-id <e>` |
| Read one record | `jodoo +data-get --app-id <a> --entry-id <e> --data-id <d>` |
| Filter / paginate records | `jodoo +data-list ... --filter <json> [--paginate-all]` |
| Create one record | `jodoo +data-create --app-id <a> --entry-id <e> --data <json>` |
| Batch create (≤100) | `jodoo +data-batch-create --data-list <json> --transaction-id <uuid>` |
| Update one record | `jodoo +data-update --data-id <d> --data <json>` |
| Batch update | `jodoo +data-batch-update --data-ids X --data-ids Y --data <json>` |
| Delete one / many | `jodoo +data-delete` / `+data-batch-delete` (high-risk → `--yes`) |
| Get upload URL/token | `jodoo +file-get-token --app-id <a> --entry-id <e> --transaction-id <uuid>` |
| Upload one file | `jodoo +file-upload --url <u> --token <t> --file ./photo.png` |
| Get workflow instance | `jodoo +workflow-instance-get --instance-id <i> --tasks-type 1` |
| Approve / reject task | `jodoo +workflow-task-forward / +workflow-task-reject` |
| Return task to a node | `jodoo +workflow-task-back --flow-id <n>` |
| Transfer task | `jodoo +workflow-task-transfer --transfer-username <u>` |
| List user tasks | `jodoo +workflow-task-list --username <u>` |
| List CC notifications | `jodoo +workflow-cc-list --username <u> --read-status all` |
| Approval comments (UI link) | `jodoo +approval-comments --app-id <a> --entry-id <e> --data-id <d>` |
| Members CRUD | `jodoo +member-list / +member-get / +member-create / +member-update / +member-delete` |
| Bulk import members | `jodoo +member-batch-import --users @users.json` |
| Departments CRUD | `jodoo +department-list / +department-create / ...` |
| Bulk import departments | `jodoo +department-batch-import --departments @tree.json --yes` |
| Roles | `jodoo +role-list / +role-create / +role-member-list` |
| Raw API escape hatch | `jodoo-cli api /v5/<path> --data '<json>'` |

## Output controls

Every shortcut accepts:

- `--format json` (default) — full envelope `{ ok, data, meta }`
- `--format pretty` — terse human format
- `--format table | csv | ndjson` — flatten arrays of objects
- `--jq <expr>` (`-q` shorthand) — gojq filter, e.g. `-q '.data.apps[].app_id'`
- `--dry-run` — print the prepared `POST <url>` + body without sending

Pipe-friendly examples:

```bash
jodoo-cli jodoo +app-list -q '.data.apps[].app_id'
jodoo-cli jodoo +data-list --app-id A --entry-id E --paginate-all -q '.data.data | length'
jodoo-cli jodoo +member-list --dept-no 1 --has-child true --format csv > members.csv
```

## Filter DSL

`+data-list --filter` accepts the Jodoo filter object verbatim:

```json
{
  "rel": "and",
  "cond": [
    { "field": "_widget_xxx", "type": "text", "method": "eq", "value": ["foo"] },
    { "field": "_widget_yyy", "type": "number", "method": "range", "value": [1, 99] }
  ]
}
```

`rel` is `and|or`. `method` covers `eq, ne, in, nin, range, like, gt, lt,
all, empty, not_empty, verified, unverified`. See the README of jodoo-cli
for the field-type → method matrix.

## File upload flow

Every file field expects a two-step dance:

1. **Get upload credentials** (returns up to 100 URL+token pairs, scoped
   to one `transaction_id`):

   ```bash
   TXN=$(uuidgen)
   jodoo-cli jodoo +file-get-token --app-id A --entry-id E --transaction-id "$TXN"
   ```

2. **Upload each file** to its paired URL:

   ```bash
   jodoo-cli jodoo +file-upload --url "$URL" --token "$TOKEN" --file ./photo.png
   ```

   The response contains a `key`. Use that key when populating the
   image / attachment field in step 3.

3. **Create / update the record** with the SAME `transaction_id`:

   ```bash
   jodoo-cli jodoo +data-create \
     --app-id A --entry-id E --transaction-id "$TXN" \
     --data '{"_widget_photo":[{"key":"<key from step 2>"}]}'
   ```

   Without the matching `transaction_id`, Jodoo rejects the upload (error
   17025). For batch retries use the same `transaction_id` to avoid
   double-inserting (the server is idempotent on it).

## Pagination

`+data-list` uses cursor pagination keyed on `data_id` (the last record's
`_id` from the previous page). Two ways to consume:

- `--paginate-all` walks the entire form, but be mindful — the multi-record
  endpoint is rate-limited to **5 requests/second**.
- Manual loop: feed the previous response's last `_id` back via
  `--data-id <id>` until the returned slice is shorter than `--limit`.

## Error handling

The CLI exits with a JSON envelope on stderr:

```json
{"ok":false,"error":{"type":"permission","code":8302,"message":"...","hint":"..."}}
```

Common codes:

- `8301 / 17018` — bad API key → `jodoo-cli auth set` or fix `JODOO_API_KEY`
- `8302` — key not authorized for the target app → re-scope on Open Platform
- `8303 / 8304` — rate-limit; back off
- `17025 / 17026` — bad / duplicate `transaction_id`
- `4815` — malformed filter object
- `7212 / 7216-7219` — quota / plan limits

The error type (`auth | permission | rate_limit | validation | quota |
api`) is stable and safe for scripts.

## Raw API escape hatch

Anything not yet wrapped:

```bash
jodoo-cli api /v5/something/new --data '{"foo":"bar"}' --jq '.data'
jodoo-cli api /v5/app/list --data @./payload.json --raw   # include code/msg
```

`--dry-run` previews; `--raw` keeps the envelope (`code`, `msg`) in
the printed JSON instead of stripping it.

## When to prefer the raw API over a shortcut

Almost never. A shortcut adds:

- `--dry-run` preview
- typed flags (`--limit` clamped, `--data-id` documented, …)
- response shaping (envelope-stripped data)
- per-shortcut tips in `--help`

Use raw API only if Jodoo released an endpoint we haven't wrapped yet —
then file an issue / PR to add a shortcut.

## Profiles

Multiple companies / API keys live side by side:

```bash
jodoo-cli config init --profile companyA --api-key sk-...
jodoo-cli config init --profile companyB --use-keychain
jodoo-cli profile list
jodoo-cli profile use companyA
jodoo-cli --profile companyB jodoo +app-list   # one-shot override
```

`JODOO_PROFILE=companyB` env override beats the file default but loses
to `--profile companyB` on the command line.
