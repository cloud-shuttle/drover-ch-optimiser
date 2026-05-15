# How to integrate with drover-code

**Problem:** You want Drover agents to run ClickHouse audits and act on structured results.

## Prerequisites

- `drover-ch` on `PATH` (built or installed)
- ClickHouse credentials available to the agent environment (read-only user)
- drover-code with shell/tool execution enabled

## JSON output for agents

```bash
drover-ch audit --format json --output /tmp/ch-audit.json
```

The JSON payload includes:

- `health_score`, issue lists, `mermaid` content
- Per-table summaries (engine, rows, issue counts)

Agents should parse `CriticalIssues` / `HighIssues` and prioritize by `rule_id` + `table`.

## Example agent workflow

1. **Discover** — agent runs `drover-ch audit -f json`
2. **Plan** — read critical issues; fetch official rule docs if needed
3. **Propose** — suggest DDL changes (never auto-apply in production)
4. **Verify** — re-run with `--focus` after human applies change

## Drover custom command

This repository ships [`.drover/commands/clickhouse-audit.md`](../../.drover/commands/clickhouse-audit.md). Copy or symlink it into your project’s `.drover/commands/` directory.

## Security

- Use a dedicated **read-only** ClickHouse user
- Do not pass passwords on the command line in agent logs; use config file or env
- Scrub reports before posting to external ticket systems if they contain table/column names you consider sensitive

## Related

- [Output formats](../reference/output-formats.md)
- [Configuration](../reference/configuration.md)
