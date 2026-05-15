# How to generate a dependency graph only

**Problem:** You want a warehouse map (MV chains, query hot paths) without running the full rules engine report.

## Mermaid to stdout

```bash
drover-ch graph --days 30 > warehouse.mmd
```

Paste into GitHub Markdown, Notion, or any Mermaid renderer.

## Self-contained HTML

```bash
drover-ch graph --format html --output warehouse-graph.html
```

## Focused subgraph

```bash
drover-ch graph --focus default.events --format html -o events-graph.html
```

## Flags

Same as `audit` for connection and scope: `--config`, `--days`, `--focus`, `--database`, `-v`.

## What appears in the graph

| Edge type | Source |
|-----------|--------|
| MV reads | `system.tables` dependencies |
| Distributed → local | Heuristic same-database name |
| Query usage | `system.query_log` (if enabled) |

Node stroke color reflects optimizer health when you ran a full audit path; graph-only HTML may show default styling.

## Related

- [Why graph visualization](../explanation/why-graph-visualization.md)
- [Architecture](../explanation/architecture.md)
