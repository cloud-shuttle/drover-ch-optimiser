# Best-practices philosophy

**Explanation** — why this tool exists and how it chooses what to flag.

## Schema beats query tuning

In ClickHouse, the largest wins usually come from:

1. **ORDER BY / primary key** matching filter patterns
2. **Data types** sized correctly (not everything as `String`)
3. **Partitioning** with bounded cardinality
4. **Ingest patterns** that avoid mutations and part explosion

Query-level tweaks matter, but after the storage layout is wrong, you are fighting the engine.

## Low-hanging fruit first

drover-ch-optimiser intentionally targets **high-impact, automatable** checks:

- Heuristics over perfect SQL parsing
- Clear severity (critical / high)
- Explicit “human review required” suggestions

It is not a replacement for a senior DBA review or load testing.

## Alignment with official rules

Rule IDs mirror the [ClickHouse agent-skills best-practices](https://github.com/ClickHouse/agent-skills/tree/main/skills/clickhouse-best-practices) catalog so that:

- Humans can cross-reference `rules/schema-*.md` docs
- Agents trained on the same skill give consistent advice
- The tool and documentation stay in sync as the catalog grows

## False positives

Expect some noise:

- Column **names** suggesting `String` abuse without seeing data distribution
- Partition hints without exact partition counts
- PK/filter rules until query_log column extraction matures

Use `--focus`, `--days`, and your domain knowledge to filter.

## Related

- [Rules catalog](../reference/rules-catalog.md)
- [Comparison with other tools](comparison-with-other-tools.md)
