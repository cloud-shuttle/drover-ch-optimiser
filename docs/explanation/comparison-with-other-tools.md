---
title: "Comparison with other tools"
description: "How drover-ch-optimiser compares with other ClickHouse auditing and observability tools."
product: drover-ch-optimiser
audience: platform-operator
doc_type: explanation
topics:
  - data-warehousing
surface: repo-docs
---

# Comparison with other tools

**Explanation** — where drover-ch-optimiser fits in the ecosystem.

## What this tool is

- **Offline / on-demand** warehouse audit
- **Rule-driven** against official best practices
- **Graph + report** in one binary
- **Agent-ready** JSON

## What it is not

- Continuous monitoring (use OTel + Grafana / ClickHouse Cloud console)
- Automatic query rewriting or schema migration
- dbt project linter (no manifest required)

## Landscape

| Tool / approach | Focus | vs drover-ch-optimiser |
|-----------------|-------|------------------------|
| [clickhouse-table-graph](https://github.com/mbaksheev/clickhouse-table-graph) | MV / table DAG | We add rules, health scores, query_log weights |
| QuerySight / similar | Query log analytics | Often dbt- or query-centric; we emphasize **schema** rules |
| Manual consultant audit | Custom | We codify repeatable checks |
| `EXPLAIN` / `system.query_log` SQL | Ad hoc | We package discovery + graph + prioritized output |
| ClickHouse Cloud advisor | Managed | drover-ch is self-hosted / portable |

## When to use drover-ch-optimiser

- Weekly health cron on self-hosted ClickHouse
- Pre-release schema review in CI
- Consulting deliverable (HTML report)
- Drover agents auditing customer warehouses

## Related

- [Best-practices philosophy](best-practices-philosophy.md)
- [Get started](../tutorials/getting-started.md)
