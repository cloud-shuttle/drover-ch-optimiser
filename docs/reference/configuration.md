# Configuration reference

## Config file

**Path (search order):**

1. `--config /path/to/file.yaml`
2. `~/.drover/ch-optimiser.yaml` (if exists)

**Example:**

```yaml
clickhouse:
  host: localhost
  port: 9000
  database: default
  username: default
  password: ""
  secure: false
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `clickhouse.host` | string | `localhost` | Server hostname |
| `clickhouse.port` | int | `9000` | Native protocol port |
| `clickhouse.database` | string | `default` | Default database |
| `clickhouse.username` | string | `default` | Username |
| `clickhouse.password` | string | `""` | Password |
| `clickhouse.secure` | bool | `false` | Enable TLS |

!!! note
    The client uses **native TCP** (`clickhouse-go`), default port **9000**. HTTP port 8123 is not used unless you terminate TLS on a proxy that speaks native protocol.

## Environment variables

| Variable | Maps to |
|----------|---------|
| `CLICKHOUSE_HOST` | `clickhouse.host` |
| `CLICKHOUSE_PORT` | `clickhouse.port` |
| `CLICKHOUSE_DATABASE` | `clickhouse.database` |
| `CLICKHOUSE_USER` | `clickhouse.username` |
| `CLICKHOUSE_PASSWORD` | `clickhouse.password` |

## ClickHouse Cloud

- Set `host` to your Cloud endpoint
- `secure: true`, port often `9440`
- Use a read-only user with `SELECT` on required databases and `system` tables

## Recommended ClickHouse grants (read-only audit)

Minimum for full features:

- `SELECT` on application databases
- `SELECT` on `system.tables`, `system.columns`, `system.parts`
- `SELECT` on `system.query_log` (optional; graph/query rules degrade gracefully)

## Security

- Passwords are not printed by the CLI
- Prefer config file permissions `0600` or secret manager injection in CI

## Related

- [CLI reference](cli-reference.md)
- [Get started tutorial](../tutorials/getting-started.md)
