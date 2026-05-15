# ClickHouse warehouse audit

Run **drover-ch-optimiser** against the configured ClickHouse cluster and return structured findings for the user or downstream agents.

## When to use

- Before changing `ORDER BY`, partition keys, or adding skipping indexes
- Periodic warehouse health checks (weekly/monthly)
- After onboarding a new database to ClickHouse Cloud

## Prerequisites

- `drover-ch` on `PATH` (see [installation](https://cloud-shuttle.github.io/drover-ch-optimiser/reference/installation/))
- Read-only credentials in `~/.drover/ch-optimiser.yaml` or env (`CLICKHOUSE_*`)
- ClickHouse Cloud: `secure: true`, port `9440`

## Steps

1. **Run audit (JSON for agents)**

   ```bash
   drover-ch audit --format json --output .drover/last-ch-audit.json
   ```

   For large warehouses, narrow scope:

   ```bash
   drover-ch audit -f json -o .drover/last-ch-audit.json \
     --database analytics \
     --exclude otel --exclude peerdb
   ```

2. **Read results** — open `.drover/last-ch-audit.json` and summarize:
   - `HealthScore` (0–100)
   - `CriticalIssues` then `HighIssues` (table, `rule_id`, message, suggestion)

3. **Optional graph** — human-readable report:

   ```bash
   drover-ch audit -f html -o .drover/clickhouse-audit.html
   ```

4. **Table-specific follow-up**

   ```bash
   drover-ch suggest <database.table> -f json -o .drover/suggest-<table>.json
   ```

   Includes impact analysis (upstream MVs, downstream dependents, query_log filter columns).

5. **Rules reference**

   ```bash
   drover-ch rules
   ```

## Agent rules

- **Never** run `ALTER TABLE`, `DROP`, or `TRUNCATE` without explicit user approval.
- Cite `rule_id` from output when recommending changes; link to [ClickHouse best practices](https://clickhouse.com/docs/best-practices).
- Treat `query_log`-derived filter columns as hints — confirm with `EXPLAIN` / production dashboards.
- Redact passwords and connection strings from chat or tickets.

## Telemetry (optional)

If your team enables usage tracking, log only: timestamp, `--database`, issue counts, health score — not raw DDL or row data.

## Related

- Repo: https://github.com/cloud-shuttle/drover-ch-optimiser
- MCP tool (v1.1): `drover-ch-mcp` — `clickhouse_audit` returns the same JSON payload
