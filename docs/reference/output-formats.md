# Output formats reference

## Overview

| Format | Flag | Default file | Best for |
|--------|------|--------------|----------|
| Markdown | `-f markdown` | `clickhouse-audit.md` | GitHub, PRs, docs |
| HTML | `-f html` | `clickhouse-audit.html` | Sharing, executives |
| JSON | `-f json` | `clickhouse-audit.json` | Agents, CI, tooling |

## Markdown

Sections:

- Title and metadata (timestamp, health score, table count)
- Critical issues (bullet per finding)
- High issues
- Mermaid code block (dependency graph)

Compatible with GitHub-flavored Markdown renderers that support Mermaid.

## HTML

Self-contained single file:

- Tailwind (CDN) layout, dark theme
- Embedded Mermaid.js (CDN)
- Issue cards by severity
- Table overview grid

Requires network when opening (CDN scripts) unless you vendor assets in a future version.

## JSON

Top-level fields (representative):

```json
{
  "Database": "default",
  "GeneratedAt": "...",
  "HealthScore": 72.5,
  "TotalTables": 42,
  "QueryDays": 30,
  "CriticalIssues": [],
  "HighIssues": [],
  "MermaidContent": "flowchart TB\n...",
  "Tables": []
}
```

Each issue object:

| Field | Type | Description |
|-------|------|-------------|
| `rule_id` | string | Rule identifier |
| `severity` | string | `critical`, `high`, … |
| `message` | string | Human-readable finding |
| `suggestion` | string | Remediation hint |
| `table` | string | `db.table` |

## Graph-only output

`drover-ch graph` without HTML:

- **stdout**: raw Mermaid `flowchart TB` source

`drover-ch graph -f html`:

- HTML wrapper with graph only (minimal report chrome)

## Related

- [CLI reference](cli-reference.md)
- [Integrate with drover-code](../how-to/integrate-with-drover-code.md)
