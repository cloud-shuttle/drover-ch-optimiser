# drover-ch-optimiser Documentation

**Intelligent ClickHouse warehouse auditor and dependency graph visualizer.**

Part of the [Drover](https://github.com/cloud-shuttle) family. Applies the official [ClickHouse best-practices](https://clickhouse.com/docs/best-practices) rules to your schemas and query history, then produces actionable reports.

## Quick start

```bash
go build -o drover-ch ./cmd/drover-ch
./drover-ch audit --format html -o audit.html
```

[Install and first run →](tutorials/getting-started.md)

## Choose your path (Diátaxis)

| | **Learning** | **Practical work** |
|---|--------------|-------------------|
| **Hands-on** | **[Tutorials](tutorials/)** — step-by-step lessons | **[How-to guides](how-to/)** — solve a specific problem |
| **Understanding** | **[Explanation](explanation/)** — background and why | **[Reference](reference/)** — technical facts |

### Tutorials

- [Get started](tutorials/getting-started.md) — install, configure, first audit
- [Your first audit](tutorials/first-audit.md) — read the report and graph

### How-to guides

- [Audit a single table](how-to/audit-specific-table.md)
- [Generate a graph only](how-to/generate-graph-only.md)
- [Integrate with drover-code](how-to/integrate-with-drover-code.md)
- [Schedule regular audits](how-to/schedule-regular-audits.md)

### Reference

- [CLI reference](reference/cli-reference.md)
- [Rules catalog](reference/rules-catalog.md)
- [Configuration](reference/configuration.md)
- [Output formats](reference/output-formats.md)

### Explanation

- [Architecture](explanation/architecture.md)
- [Best-practices philosophy](explanation/best-practices-philosophy.md)
- [Why graph visualization](explanation/why-graph-visualization.md)
- [Comparison with other tools](explanation/comparison-with-other-tools.md)

## Also see

- [DESIGN.md](https://github.com/cloud-shuttle/drover-ch-optimiser/blob/main/DESIGN.md) — implementation design (repo root)
- [`.beads/issues.jsonl`](https://github.com/cloud-shuttle/drover-ch-optimiser/blob/main/.beads/issues.jsonl) — project backlog
