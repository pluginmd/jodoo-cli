# jodoo-cli

A focused CLI for **Jodoo** (`api.jodoo.com`) — built in Go, modeled on the
`basecli` architecture.

- 30+ curated `jodoo` shortcuts — apps, forms, records, files, workflow,
  contacts (members / departments / roles)
- Raw API escape hatch (`jodoo-cli api <path>`) for anything not yet wrapped
- Bearer-token auth, profiles, doctor, completion, output formats
  (json / pretty / table / ndjson / csv) with `--jq` filtering
- Two AI agent skills under `skills/working/`:
  `jodoo-cli`, `jodoo-shared`

## Install from source

Requires Go `v1.23+`.

```bash
make install              # builds and installs to /usr/local/bin/jodoo-cli
# or just
make build && ./jodoo-cli --help
```

## First-time setup

```bash
# 1. Configure your API key (interactive)
jodoo-cli config init

# 2. Verify
jodoo-cli doctor
```

You can also export `JODOO_API_KEY` in your environment to skip the config
step (useful for CI).

## Common commands

```bash
# List apps reachable by your API key
jodoo-cli jodoo +app-list

# List forms in an app
jodoo-cli jodoo +form-list --app-id <app_id>

# Inspect form fields
jodoo-cli jodoo +widget-list --app-id <app_id> --entry-id <entry_id>

# Read a single record
jodoo-cli jodoo +data-get --app-id <a> --entry-id <e> --data-id <d>

# Filter / paginate records
jodoo-cli jodoo +data-list --app-id <a> --entry-id <e> \
  --filter '{"rel":"and","cond":[{"field":"_widget_xxx","type":"text","method":"eq","value":["foo"]}]}' \
  --limit 100

# Create a record
jodoo-cli jodoo +data-create --app-id <a> --entry-id <e> \
  --data '{"_widget_1529400746031":"Hello Jodoo"}'

# Batch create
jodoo-cli jodoo +data-batch-create --app-id <a> --entry-id <e> \
  --data-list '[{"_widget_xxx":"a"},{"_widget_xxx":"b"}]'

# Workflow
jodoo-cli jodoo +workflow-instance-get --instance-id <id> --tasks-type 1
jodoo-cli jodoo +workflow-task-list   --username <user>
jodoo-cli jodoo +workflow-task-forward --username <user> --instance-id <id> --task-id <task>

# Contact
jodoo-cli jodoo +member-list --dept-no 1 --has-child true
jodoo-cli jodoo +member-create --name "Alice" --username alice
jodoo-cli jodoo +department-list --dept-no 1

# Files (two-step upload)
jodoo-cli jodoo +file-get-token --app-id <a> --entry-id <e> --transaction-id <uuid>
jodoo-cli jodoo +file-upload --url <upload_url> --token <token> --file ./photo.png
```

## Escape hatch: raw API

Anything not covered by a shortcut can still be reached via the raw API.
Jodoo APIs are all `POST` to `https://api.jodoo.com/api`:

```bash
jodoo-cli api /v5/app/list --data '{"limit":100,"skip":0}'
jodoo-cli api /v5/app/entry/data/list \
  --data '{"app_id":"...","entry_id":"...","limit":50}'
```

Use `--dry-run` to preview the request without sending it.

## Command layers

```
jodoo-cli api      # raw POST endpoint
jodoo-cli jodoo    # 30+ curated shortcuts (human & agent friendly)
jodoo-cli auth     # set / show / clear API key
jodoo-cli config   # API key, base URL, profiles
jodoo-cli profile  # switch / rename / remove profiles
jodoo-cli doctor   # health check (config, auth, connectivity)
```

## Project layout

```
cmd/                # cobra commands: api, auth, config, doctor, profile, root
internal/           # plumbing: build, client, cmdutil, core, credential,
                    # output, validate
shortcuts/          # jodoo/ (~30 shortcuts) + common/ + register.go
skills/working/     # AI agent skills (isolated workspace)
docs/               # research.md (Jodoo API specs)
```

