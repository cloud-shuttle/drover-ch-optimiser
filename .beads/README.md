# Beads backlog — drover-ch-optimiser

Agent-ready issue tracker for this repo. Issues live in `issues.jsonl` (one JSON object per line).

## Quick commands

```bash
bd ready --json          # unblocked work
bd start drover-ch-005   # claim a task
bd update drover-ch-005 --status in_progress
bd close drover-ch-004   # mark done
```

## Status snapshot (starter import)

| ID | Title | Status |
|----|-------|--------|
| drover-ch-001 | Core CLI epic | closed |
| drover-ch-002 | Connection + config | closed |
| drover-ch-003 | Graph builder | closed |
| drover-ch-004 | Rules engine | in_progress |
| drover-ch-005 | Query log enrichment | in_progress |
| drover-ch-006 | HTML report | closed |
| drover-ch-007 | Markdown + JSON | closed |
| drover-ch-008 | Drover integration | open |
| drover-ch-009 | Test suite | open |
| drover-ch-010 | Docs + v0.1.0 release | in_progress |
| drover-ch-011–013 | v1.1 epics | open |

## Drover sync

The main [Drover](https://github.com/cloud-shuttle/drover) repo uses `.beads/beads.jsonl` (event log format). This project uses standalone `issues.jsonl` for `bd` until wired into Drover export/import.
