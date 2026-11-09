---
title: "How to schedule regular audits"
description: "Run audits on a schedule (cron) and keep the HTML reports."
product: drover-ch-optimiser
audience: platform-operator
doc_type: how-to
topics:
  - data-warehousing
surface: repo-docs
---

# How to schedule regular audits

**Problem:** Run warehouse health checks on a schedule (cron, CI, Kubernetes Job) and retain history.

## Cron example

```bash
# Weekly Sunday 02:00 — HTML report
0 2 * * 0 /usr/local/bin/drover-ch audit -f html -o /var/reports/ch-audit-$(date +\%Y-\%m-\%d).html
```

## CI (GitHub Actions sketch)

```yaml
jobs:
  clickhouse-audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: go build -o drover-ch ./cmd/drover-ch
      - name: Audit
        env:
          CLICKHOUSE_HOST: ${{ secrets.CH_HOST }}
          CLICKHOUSE_USER: ${{ secrets.CH_AUDIT_USER }}
          CLICKHOUSE_PASSWORD: ${{ secrets.CH_PASSWORD }}
        run: ./drover-ch audit -f json -o audit.json
      - uses: actions/upload-artifact@v4
        with:
          name: clickhouse-audit
          path: audit.json
```

## Fail CI on critical issues (optional)

```bash
drover-ch audit -f json -o audit.json
CRITICAL=$(jq '[.CriticalIssues[]] | length' audit.json)
if [ "$CRITICAL" -gt 0 ]; then
  echo "Found $CRITICAL critical ClickHouse issues"
  exit 1
fi
```

Adjust policy to your tolerance — many teams report-only on first adoption.

## Kubernetes Job

Run a minimal image with `drover-ch` binary; mount config Secret at `/root/.drover/ch-optimiser.yaml`; write reports to S3 via sidecar or post-job upload.

## Performance tips

- Use `--database` to shard audits per database on very large clusters
- Reduce `--days` for faster query_log scans
- Run during off-peak hours

## Related

- [Configuration](../reference/configuration.md)
- [Output formats](../reference/output-formats.md)
