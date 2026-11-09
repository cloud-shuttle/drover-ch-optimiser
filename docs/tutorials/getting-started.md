# Tutorial: Get started with drover-ch-optimiser

**Goal:** Run your first warehouse audit in under five minutes.

This is a **tutorial** — it teaches by doing. For command flags see [CLI reference](../reference/cli-reference.md). For why the tool works this way see [Architecture](../explanation/architecture.md).

## Prerequisites

- Go 1.23 or newer
- A ClickHouse server (local Docker, Cloud, or self-hosted)
- A user with **read-only** access to `system.*` and your application databases

## Step 1: Build or install

From the repository:

```bash
git clone https://github.com/cloud-shuttle/drover-ch-optimiser.git
cd drover-ch-optimiser
go build -o drover-ch ./cmd/drover-ch
```

Or install directly (when published):

```bash
go install github.com/cloud-shuttle/drover-ch-optimiser/cmd/drover-ch@latest
```

Verify:

```bash
./drover-ch --help
./drover-ch rules    # no ClickHouse required
```

## Step 2: Configure connection (optional)

Defaults: `localhost:9000`, database `default`, user `default`.

**Option A — config file** (`~/.drover/ch-optimiser.yaml`):

```yaml
clickhouse:
  host: localhost
  port: 9000          # native protocol (default)
  database: default
  username: default
  password: ""
  secure: false
```

Copy from [`.drover/ch-optimiser.example.yaml`](https://github.com/cloud-shuttle/drover-ch-optimiser/blob/main/.drover/ch-optimiser.example.yaml).

**Option B — environment variables:**

```bash
export CLICKHOUSE_HOST=ch.example.com
export CLICKHOUSE_PORT=9440
export CLICKHOUSE_USER=audit_reader
export CLICKHOUSE_PASSWORD=secret
```

See [Configuration](../reference/configuration.md) for TLS and Cloud notes.

## Step 3: Run your first audit

```bash
./drover-ch audit --format html --output my-audit.html -v
```

- `-v` prints discovery and analysis progress
- Open `my-audit.html` in a browser

You should see:

1. **Health score** (0–100) for the warehouse
2. **Critical / high issues** with rule IDs and suggestions
3. **Mermaid dependency graph** (tables, MVs, query hot paths)
4. **Table overview** with per-table health

## Step 4: Try other formats

```bash
./drover-ch audit -f markdown -o audit.md
./drover-ch audit -f json -o audit.json
```

JSON is ideal for agents and CI parsers.

## What you learned

- How to build and invoke `drover-ch`
- How connection config works
- How to produce HTML, Markdown, and JSON audits

## Next

→ [Tutorial: Your first audit](first-audit.md) — interpret scores, graphs, and issues.
