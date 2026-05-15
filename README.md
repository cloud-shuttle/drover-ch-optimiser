# drover-ch-optimiser

ClickHouse warehouse auditor and dependency-graph visualizer — part of the [Drover](https://github.com/cloud-shuttle) family.

Runs the official [ClickHouse best-practices](https://clickhouse.com/docs/best-practices) rules against your schemas and query history, then produces Markdown, HTML, or JSON reports with an interactive Mermaid graph.

## Install

| Method | Command |
|--------|---------|
| Go | `go install github.com/cloud-shuttle/drover-ch-optimiser/cmd/drover-ch@latest` |
| Release | [GitHub Releases](https://github.com/cloud-shuttle/drover-ch-optimiser/releases) — `drover-ch` binaries for linux/darwin/windows |
| Homebrew | `brew install ./packaging/homebrew/drover-ch.rb` (update sha after first tag) |

Docs: **[cloud-shuttle.github.io/drover-ch-optimiser](https://cloud-shuttle.github.io/drover-ch-optimiser/)** (GitHub Pages from `main`).

## Testing

```bash
# Unit tests (no Docker) — includes golden HTML report diff
make test
make test-cover

# Regenerate golden report after template changes
go test ./internal/report -update

# Integration + e2e (requires Docker)
make ch-up
make test-integration
make test-e2e
make test-all
make ch-down
```

CI runs unit, integration, and e2e jobs on every push (`.github/workflows/test.yml`).  
Integration tests use ClickHouse on port **19000** (`docker-compose.test.yml`) and fixtures in `testdata/clickhouse/init.sql`.

## Documentation

Full [Diátaxis](https://diataxis.fr/)-structured docs in [`docs/`](docs/index.md).

```bash
make docs-serve    # mkdocs serve
make docs-build    # site/ output
make docs-deploy   # mkdocs gh-deploy (maintainers)
```

## Quick start

```bash
go build -o drover-ch ./cmd/drover-ch

# Configure (~/.drover/ch-optimiser.yaml or env vars)
export CLICKHOUSE_HOST=localhost
export CLICKHOUSE_PORT=9000

./drover-ch audit --format html -o audit.html
./drover-ch graph --focus default.events
./drover-ch rules
```

### Sample config

```yaml
# ~/.drover/ch-optimiser.yaml
clickhouse:
  host: localhost
  port: 9000
  database: default
  username: default
  password: ""
  secure: false
```

## Commands

| Command | Description |
|---------|-------------|
| `audit` | Full scan + rules + report |
| `graph` | Dependency / query graph only |
| `rules` | List embedded rule IDs |
| `suggest [db.table]` | Focused audit on one table |

### Flags

- `--format` / `-f`: `markdown` (default), `html`, `json`
- `--output` / `-o`: output file path
- `--focus`: limit graph to a table and its edges
- `--days`: `query_log` lookback (default 30)
- `--database`: scan a single database
- `--exclude`: skip databases/tables (repeatable; e.g. `otel`, `default.temp_*`, `peerdb*`)

## Project tracking

Backlog lives in [`.beads/issues.jsonl`](.beads/issues.jsonl) (Beads / `bd`).

```bash
bd ready --json
bd start drover-ch-005
```

See [`.beads/README.md`](.beads/README.md) for current status.

## Architecture

See [DESIGN.md](./DESIGN.md) for the full design document.

```
cmd/drover-ch          CLI
internal/chclient      ClickHouse connection
internal/graph         Discovery, edges, Mermaid
internal/analyzer      YAML rules + evaluators
internal/report        Markdown / HTML / JSON
```

## Requirements

- Go 1.23+
- ClickHouse with read access to `system.*` and target databases
- `system.query_log` optional (enables query-weighted edges)

## License

Apache-2.0 (align with ClickHouse best-practices skill)
