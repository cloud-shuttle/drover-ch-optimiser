---
title: "Architecture"
description: "How drover-ch-optimiser is structured: collectors, analyzer rules, graph and reporters."
product: drover-ch-optimiser
audience: platform-operator
doc_type: explanation
topics:
  - data-warehousing
surface: repo-docs
---

# Architecture

**Explanation** — how the system is structured and why.

For implementation detail see [DESIGN.md](https://github.com/cloud-shuttle/drover-ch-optimiser/blob/main/DESIGN.md). For commands see [CLI reference](../reference/cli-reference.md).

## Layered pipeline

```
CLI (Cobra)
    → config + chclient (ClickHouse connection)
    → graph.BuildGraph (discovery + edges)
    → analyzer.Run (rules, parallel per table)
    → report (Markdown / HTML / JSON)
```

## Components

### Connection layer (`internal/chclient`)

- `clickhouse-go/v2` native protocol
- Ping on connect, `max_execution_time` cap
- Read-only usage by convention

### Discovery layer (`internal/graph`)

Queries:

- `system.tables` — engines, keys, dependencies, row stats
- `system.columns` — types for rule evaluation
- `system.parts` — active part counts
- `system.query_log` — optional query frequency / duration edges

Builds an in-memory `Graph` of nodes and directed edges.

### Rules engine (`internal/analyzer`)

- Rules declared in embedded YAML
- Go `switch` handlers per `condition.type`
- Parallel evaluation per table (`sync.WaitGroup`)
- Warehouse and per-table health scores

### Report layer (`internal/report`)

- Go `html/template` for HTML
- String builder for Markdown
- `encoding/json` for agents

## Design principles

| Principle | Rationale |
|-----------|-----------|
| Static binary | Easy cron, CI, consultant laptop |
| Read-only | Safe on production |
| Pluggable rules | YAML + Go for complex checks |
| Graph + rules together | Context for impact (MVs, hot queries) |

## Extension points

- New rules: add YAML + handler in `evaluator.go` / `helpers.go`
- New edge types: extend `graph.EdgeType` and builder
- New output: implement writer on `report.Report`

## Related

- [Why graph visualization](why-graph-visualization.md)
- [Best-practices philosophy](best-practices-philosophy.md)
