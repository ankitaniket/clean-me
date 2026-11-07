# clean-me

A safe, multi-stage system cleanup CLI. It reclaims disk space across **Docker, OS caches, Node.js and Python projects**, and audits **stale git branches** and **Redis memory** — but it never deletes blindly.

Every item is previewed with its **path · purpose · reclaimable size**. Deletions go to the **trash first** (recoverable). Dangerous operations require an explicit `--force`. Safe to run on production servers and professional laptops.

```
▸ NODE  (7.0 GB reclaimable)
  TIER  SIZE    PURPOSE                              PATH
  safe  7.0 GB  npm download cache, re-downloadable  ~/.npm/_cacache

▸ SYSTEM  (299.4 MB reclaimable)
  TIER  SIZE      PURPOSE                                            PATH
  safe  299.4 MB  old formula downloads & versions, re-downloadable  Homebrew cache

Total reclaimable: 7.3 GB
```

## Philosophy

clean-me does not reinvent cleanup logic. It is an **orchestrator + safety layer** on top of tools that are already proven: `docker system df`/`prune`, `brew`/`apt`/`dnf`, `npm`/`pip` caches, and `git`. clean-me's job is to make them **safe to run everywhere** by always answering three questions before it touches anything:

1. **What** exactly will be removed (the path)?
2. **Why** is it safe to remove (the purpose)?
3. **How much** space does it free (the reclaimable size)?

## Install

```sh
go install github.com/ankitaniket/clean-me@latest
# or build a single static binary
go build -o clean-me .
```

## Usage

```sh
clean-me scan                      # read-only preview of everything (deletes nothing)
clean-me scan --json               # machine-readable output
clean-me clean                     # reclaim, asking before each item (default No)
clean-me clean --dry-run           # show the exact commands without running them
clean-me clean --stage node,python --path ~/code
clean-me clean --tier 2            # also include opt-in items (e.g. all unused Docker images)
clean-me clean --tier 3 --force    # include danger-tier (Docker volumes, virtualenvs)
clean-me clean --no-trash          # permanently delete instead of trashing
```

### Safety model

- **Dry-run first**: `scan` and `clean --dry-run` never modify anything.
- **Trash-first**: confirmed deletions move to the OS trash; empty it to finalize.
- **Risk tiers**: `safe` (tier 1, default) → `opt-in` (tier 2) → `danger` (tier 3, needs `--force`).
- **No-TTY guard**: in a non-interactive shell nothing is deleted unless `--yes` is passed.
- **Root guard**: refuses to run as root without `--allow-root`.

## Stages

| Stage | What it does | Deletes? |
|-------|--------------|----------|
| **docker** | Previews reclaimable from `docker system df`; prunes build cache, stopped containers, unused images (tier 2), unused volumes (tier 3) | yes (permanent — Docker has no trash) |
| **system** | Homebrew / apt / dnf package caches, systemd journal vacuum | yes (permanent) |
| **node** | `node_modules` (trash-first; pnpm/bun link-aware sizing) + npm cache | yes (trash-first) |
| **python** | `__pycache__`, tool caches, pip cache, virtualenvs (tier 3) | yes (trash-first) |
| **git** | Audits merged / upstream-gone / stale (>90d) branches across repos | **report-only** |
| **redis** | Audits memory usage and biggest keys (read-only, non-blocking) | **report-only** |

Git branches and Redis keys are **audit-only** by design: both are too easy to break by deleting blindly, so clean-me shows you what to review and leaves the decision to you.

## License

MIT
