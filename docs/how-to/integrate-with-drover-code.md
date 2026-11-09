---
title: "How to integrate with drover-code"
description: "Use drover-ch audit output as context for drover-code agents via the bundled command."
product: drover-ch-optimiser
audience: platform-operator
doc_type: how-to
topics:
  - data-warehousing
surface: repo-docs
---

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

This repository ships [`.drover/commands/clickhouse-audit.md`](https://github.com/cloud-shuttle/drover-ch-optimiser/blob/main/.drover/commands/clickhouse-audit.md). Copy or symlink it into your project’s `.drover/commands/` directory.

## Security

- Use a dedicated **read-only** ClickHouse user
- Do not pass passwords on the command line in agent logs; use config file or env
- Scrub reports before posting to external ticket systems if they contain table/column names you consider sensitive

## Related

- [Output formats](../reference/output-formats.md)
- [Configuration](../reference/configuration.md)
