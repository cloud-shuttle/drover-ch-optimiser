# Tutorial: Your first audit

**Goal:** Understand what the audit report tells you and what to do next.

Prerequisite: [Get started](getting-started.md).

## Run the audit

```bash
./drover-ch audit --format html --output audit.html --days 30 -v
```

Use `--days` to control how much `system.query_log` history feeds query-weighted edges (when enabled on your cluster).

## Read the health score

The headline **health score** is a weighted penalty model:

- Each **critical** issue reduces the warehouse score significantly
- **High** issues reduce it moderately
- Per-table scores appear in the table overview

A score of 100 means no rules fired. Scores in the 60–80 range often mean a few high-impact schema fixes would help a lot.

!!! tip
    Treat the score as a **prioritization aid**, not a production SLO. Always validate suggestions against your workload.

## Read priority issues

Each issue card includes:

| Field | Meaning |
|-------|---------|
| **Rule ID** | Maps to [rules catalog](../reference/rules-catalog.md) and official ClickHouse skill rules |
| **Message** | What was detected on this table |
| **Suggestion** | Typical remediation (human review required) |
| **Table** | `database.table` |

Example rule IDs:

- `schema-pk-cardinality-order` — ORDER BY column order
- `schema-types-native-types` — String used where a native type fits
- `schema-parts-too-many-active` — merge pressure from too many parts

## Read the dependency graph

The Mermaid graph shows:

- **Nodes** — tables, materialized views, views (shape varies by engine)
- **Edges** — MV dependencies, distributed→local, query hot paths
- **Colors** — node stroke reflects worst issue severity on that table

Use the graph to answer:

- Which MVs depend on a problematic raw table?
- Which tables receive the most query traffic?
- Where is complexity concentrated?

## Take action (safely)

1. Pick **critical** issues on **high-traffic** tables (cross-check with graph weights).
2. Read the linked best-practice rule in the [official docs](https://clickhouse.com/docs/best-practices).
3. Prototype fixes on a copy or new table — ORDER BY is often **immutable** without rewrite.
4. Re-run: `./drover-ch audit --focus db.table -f html`

## Common first-audit findings

| Finding | Typical fix |
|---------|-------------|
| String `created_at` | `DateTime64` |
| High-cardinality column first in ORDER BY | Reorder keys low→high cardinality |
| Many active parts | Larger insert batches, tune merges |
| Nullable everywhere | `DEFAULT` + non-nullable types |

## Next steps

- [How to audit one table](../how-to/audit-specific-table.md)
- [Rules catalog](../reference/rules-catalog.md)
- [Why graph visualization matters](../explanation/why-graph-visualization.md)
