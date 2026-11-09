---
title: "Rules catalog"
description: "Catalog of the ClickHouse best-practice rules drover-ch checks."
product: drover-ch-optimiser
audience: platform-operator
doc_type: reference
topics:
  - data-warehousing
surface: repo-docs
---

# Rules catalog

Rules are embedded in `internal/analyzer/rules.yaml` and aligned with the official [ClickHouse best-practices skill](https://github.com/ClickHouse/agent-skills/tree/main/skills/clickhouse-best-practices) (**31 rule stems** plus one Drover extension: **`drover-wide-table`** → **32** YAML entries).

Each rule has an `official_rule` field matching the skill file name (without `.md`). Audit JSON issues include **`official_rule_id`** (same value) next to `rule_id` for agents and docs cross-links.

```bash
drover-ch rules
```

Columns: **OFFICIAL** (skill stem), **ID** (stable drover id, usually identical), **SEVERITY**, **NAME**.

## Implemented vs manual

| Kind | Count | Meaning |
|------|-------|---------|
| **Static checks** | 19 | Evaluated from `system.*`, DDL, and `query_log` heuristics (incl. unbounded CROSS JOIN signal + wide-table column count) |
| **manual_review** | 13 | Listed in `drover-ch rules` with skill pointer; no automatic signal in static audit |

Manual rows cover JOIN tuning, async/format insert choices, partition “start without”, refreshable MVs, and **agent** workflow rules — cite `rules/<official_rule>.md` when reviewing manually.

## Condition types (implementation)

| Type | Typical signal |
|------|----------------|
| `sorting_key_cardinality` | ORDER BY first column vs heuristics |
| `missing_filter_in_pk` | `query_log` filter columns vs ORDER BY prefix |
| `string_abuse` | String-typed dates / amounts / IPs in DDL |
| `minimize_bitwidth` | Int128 / UInt256-style widths |
| `enum_string_candidate` | `status` / `kind` / … plain String |
| `missing_lowcardinality` | Low-cardinality string columns |
| `excessive_nullable` | Nullable column count |
| `partition_cardinality` | Partition expression heuristics |
| `merge_tree_partition_no_ttl` | Partitioned MergeTree without `TTL` |
| `partition_yyyymmdd_churn` | `toYYYYMMDD` in partition key |
| `json_dynamic_column` | `JSON` / `Object('json')` column types |
| `missing_skipping_index` | Hot filters vs DDL indexes |
| `mv_incremental_heuristic` | MV with `GROUP BY` into non-aggregating target |
| `mutation_heavy` | Incomplete `system.mutations` |
| `incomplete_delete_mutation` | Incomplete mutations whose command mentions DELETE |
| `optimize_merge_pressure` | Many active parts and/or frequent `FINAL` in `query_log` |
| `insert_batch_rows_too_small` | High insert count with low `written_rows` average |
| `cross_join_unbounded` | `query_log`: `CROSS JOIN` with no `WHERE`/`PREWHERE` in the query text (heuristic) |
| `too_many_columns` | Table has more than 80 columns (Drover extension `drover-wide-table`) |
| `manual_review` | Never emits issues; documents skill-only rules |

## Official references

- [ClickHouse best practices docs](https://clickhouse.com/docs/best-practices)
- Skill rule files: `rules/<official_rule>.md` in the agent-skills repository

## Roadmap

Further JOIN heuristics (algorithm choice, ANY, null handling): see [v1.2 follow-up](https://github.com/cloud-shuttle/drover-ch-optimiser/blob/main/.beads/issues.jsonl) `drover-ch-020` in Beads. Shipped in v1.2 (`drover-ch-019`, closed): `query_log` CROSS JOIN triage, **`drover-wide-table`** extension rule.
