# CLI reference

Binary name: **`drover-ch`**  
Version: **0.1.0**

## Global flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | | (optional file) | Path to YAML config; else `~/.drover/ch-optimiser.yaml` if present |
| `--host` | | | ClickHouse host (overrides config) |
| `--port` | | `9000` | Native protocol port |
| `--user` | | `default` | Username (overrides config) |
| `--password` | | | Password (overrides config) |
| `--secure` | | `false` | TLS (required for ClickHouse Cloud) |
| `--verbose` | `-v` | `false` | Progress logging |
| `--version` | | | Print version |
| `--help` | `-h` | | Help |

`rules` does not require a working ClickHouse connection if config file is missing.

## `drover-ch audit`

Full warehouse scan: discovery → graph → rules → report.

```bash
drover-ch audit [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | `markdown` | `markdown`, `html`, or `json` |
| `--output` | `-o` | auto-named | Output file path |
| `--focus` | | | Subgraph center `database.table` |
| `--days` | | `30` | `query_log` lookback days |
| `--database` | | | Limit to one database |

Default output files: `clickhouse-audit.md`, `.html`, or `.json`.

## `drover-ch graph`

Dependency graph only (no issue narrative beyond node styling from prior rules if embedded).

```bash
drover-ch graph [flags]
```

Supports the same flags as `audit` for format, output, focus, days, database.

- Without `--format html` and without `-o`: prints Mermaid to **stdout**
- With `-f html` or `-o`: writes `clickhouse-graph.html` (or chosen path)

## `drover-ch rules`

List embedded best-practice rules.

```bash
drover-ch rules
```

No connection required.

## `drover-ch suggest`

Focused audit for one table (sets `--focus`).

```bash
drover-ch suggest <database.table> [audit flags...]
```

Example:

```bash
drover-ch suggest analytics.events -f html -o events.html
```

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error (connection, query, or write failure) |

## Environment variables

See [Configuration](configuration.md).
