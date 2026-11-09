---
title: "Why graph visualization"
description: "Why warehouse-wide dependency graphs matter: one bad table design ripples into every downstream view and query."
product: drover-ch-optimiser
audience: platform-operator
doc_type: explanation
topics:
  - data-warehousing
surface: repo-docs
---

# Why graph visualization

**Explanation** — why a dependency graph belongs in a schema auditor.

## Tables live in a system

A bad `ORDER BY` on a raw events table does not only affect that table. It affects:

- Materialized views that read from it
- Aggregating tables downstream
- Queries that join through the chain

A flat issue list hides **blast radius**.

## What the graph shows

| Element | Insight |
|---------|---------|
| MV → source edges | Which pipelines to retest after schema change |
| Query-weighted edges | Which nodes are hot in real traffic |
| Node health color | Where rules fired hardest |
| Focus subgraph | Zoom one table’s neighborhood |

## Mermaid as default

Mermaid keeps the binary small:

- No bundled JS build pipeline
- Embeds in Markdown reports and HTML
- Familiar to GitHub and many doc tools

Future versions may add JSON export for Cytoscape / React Flow dashboards.

## When the graph is enough alone

Use `drover-ch graph` when:

- Onboarding engineers to an unfamiliar warehouse
- Documenting MV topology for a design review
- Pairing with an audit report for narrative + map

## Related

- [How to generate graph only](../how-to/generate-graph-only.md)
- [Architecture](architecture.md)
