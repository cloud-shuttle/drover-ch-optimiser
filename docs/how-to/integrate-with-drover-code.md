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

## Drover custom command (recommended)

Create in your project:

`.drover/commands/clickhouse-audit.md`:

```markdown
# ClickHouse warehouse audit

Run drover-ch-optimiser and return JSON results.

## Steps

1. Run: `drover-ch audit --format json --output .drover/last-ch-audit.json`
2. Summarize critical and high issues by table
3. For each issue, cite rule_id and suggestion
4. Do not execute ALTER TABLE without explicit user approval
```

## Security

- Use a dedicated **read-only** ClickHouse user
- Do not pass passwords on the command line in agent logs; use config file or env
- Scrub reports before posting to external ticket systems if they contain table/column names you consider sensitive

## Related

- [Output formats](../reference/output-formats.md)
- [Configuration](../reference/configuration.md)
