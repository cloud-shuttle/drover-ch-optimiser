# How to audit a single hot table

**Problem:** You need focused recommendations for one high-traffic table without noise from the whole warehouse.

## Command

```bash
drover-ch suggest default.events
# equivalent:
drover-ch audit --focus default.events --format html --output events-audit.html
```

## Useful flags

| Flag | Example | Purpose |
|------|---------|---------|
| `--focus` | `analytics.events` | Subgraph + table-scoped analysis |
| `--days` | `7` | Shorter query_log window for recent workload |
| `--database` | `analytics` | Limit discovery to one database |
| `-f html` | | Shareable report |

```bash
drover-ch audit --focus analytics.events --days 7 -f html -o events.html
```

## When to use this

- Validating a schema change before migration
- Investigating slow queries on one fact table
- Consulting deliverable for a single problematic table

## Interpret results

1. Check **critical** rules on that table first.
2. Compare **ORDER BY** / partition key to filters you know are hot.
3. Use the **graph** section to see upstream MVs and downstream consumers.

## Limitations (v0.1)

- Filter-column detection from `query_log` is still improving; some PK/filter rules need manual cross-check.
- `--focus` trims the graph to edges touching that table; it does not skip scanning other tables for warehouse score.

## Related

- [Generate graph only](generate-graph-only.md)
- [CLI reference](../reference/cli-reference.md)
