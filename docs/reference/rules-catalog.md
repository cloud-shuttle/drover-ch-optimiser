# Rules catalog

Rules are embedded in `internal/analyzer/rules.yaml` and aligned with the official [ClickHouse best-practices skill](https://github.com/ClickHouse/agent-skills/tree/main/skills/clickhouse-best-practices).

List live rules:

```bash
drover-ch rules
```

## Schema

| Rule ID | Severity | Impact | Summary | Suggestion (abbrev.) |
|---------|----------|--------|---------|----------------------|
| `schema-pk-cardinality-order` | critical | 10 | ORDER BY should be low→high cardinality | Reorder keys; recreate table if needed |
| `schema-pk-prioritize-filters` | critical | 9 | Hot filters should be in ORDER BY prefix | Add to ORDER BY or skipping index |
| `schema-types-native-types` | critical | 9 | Avoid String for dates, numbers, IPs | Use native types |
| `schema-types-lowcardinality` | high | 8 | Low-card strings → LowCardinality | `LowCardinality(String)` |
| `schema-types-avoid-nullable` | high | 7 | Nullable overhead | Prefer DEFAULT |
| `schema-partition-low-cardinality` | high | 8 | ~100–1k partitions target | `toYYYYMM` / `toDate` |
| `schema-parts-too-many-active` | high | 8 | Too many active parts | Batch inserts, merges, OPTIMIZE |

## Query

| Rule ID | Severity | Impact | Summary | Suggestion (abbrev.) |
|---------|----------|--------|---------|----------------------|
| `query-index-skipping-indices` | high | 7 | Frequent non-PK filters need indexes | ADD INDEX minmax / bloom_filter |

## Insert / mutations

| Rule ID | Severity | Impact | Summary | Suggestion (abbrev.) |
|---------|----------|--------|---------|----------------------|
| `insert-mutation-avoid-update` | critical | 10 | Heavy mutations are costly | ReplacingMergeTree, append-only |

## Condition types (implementation)

| Type | Used by |
|------|---------|
| `sorting_key_cardinality` | PK order |
| `missing_filter_in_pk` | Filter vs ORDER BY |
| `string_abuse` | Native types |
| `missing_lowcardinality` | LowCardinality |
| `excessive_nullable` | Nullable count |
| `partition_cardinality` | Partition key |
| `missing_skipping_index` | Skipping indexes |
| `mutation_heavy` | Mutations (evolving) |
| `too_many_parts` | `system.parts` |

## Roadmap

v1.1 targets the full 31-rule skill catalog. Track progress in [`.beads/issues.jsonl`](../../.beads/issues.jsonl) (`drover-ch-012`).

## Official references

- [ClickHouse best practices docs](https://clickhouse.com/docs/best-practices)
- Per-rule markdown files under `rules/` in the agent-skills repository
