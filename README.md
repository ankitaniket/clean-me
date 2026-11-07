# clean-me

**Your interactive system cleanup companion.**

A safe, interactive terminal application that helps developers discover, understand and reclaim wasted disk space. Scan your machine, explore cleanup opportunities, review exactly what will be removed, and clean with confidence.

Run it with no arguments and it launches a dashboard — no command memorization required. Think `lazygit`/`k9s`, for disk hygiene.

## The philosophy

No blind deletion. No mysterious cleanup commands. Every operation explains **what** it removes, **why** it is safe, and **how much** space it can recover. Deletions move to the trash first (recoverable). Dangerous operations require explicit confirmation. Safe to run on production servers and professional laptops.

## Install

```sh
go install github.com/ankitaniket/clean-me/cmd/clean-me@latest
# or build a single static binary
go build -o clean-me ./cmd/clean-me
```

## Interactive dashboard

```sh
clean-me
```

The dashboard is calm and read-only-first: it never scans or deletes until you ask it to.

It opens on a richer **Welcome** screen that reassures you nothing is touched until you review and confirm — "it only looks when asked." Press `Enter` to start the scan, or `?` at any time for a help overlay listing every key.

```
  clean-me

  A calm way to reclaim disk space.

  This is read-only. Nothing is scanned or deleted
  until you ask. You will review every item and
  confirm before anything is removed.

  It only looks when asked.

  ▸ Press Enter to scan      q  quit
```

While you browse, a persistent green ribbon stays pinned to the top:

```
  read-only — nothing is deleted until you confirm
```

The browsing screens are deliberately compact — **one line per item** with a detail panel for whatever is focused, so there is far less text on screen:

1. **Welcome** — read-only reassurance; `Enter` starts the scan, `?` opens help.
2. **Dashboard** — total reclaimable space, opportunity count, per-stage status, and a disk-usage bar showing used vs. free space on the volume.
3. **Categories** — Docker, Node.js, Python, System, Xcode, Caches & Trash, plus Git & Redis audits.
4. **Explorer** — one compact line per item (`[x]` · size · name) with a detail panel showing purpose, path, and risk tier for the focused item. Long lists scroll; `s` cycles the sort order (size / name) and `/` filters by a search term. Danger items are never preselected.
5. **Review** — lists **every** item that will be removed, grouped by stage.
6. **Progress** — live per-item results and the **actual** space freed.

A read-only **Space Explorer**, reachable from the dashboard, gives an `ncdu`-style drill-down of the largest directories under your home (or `--min-size`'s root), so you can understand where space has gone before deciding what to clean. It never deletes anything — it is purely for exploration.

**Selection is explicit.** Each item carries an `[x]` / `[ ]` checkbox: `Space` toggles the focused item, `a` selects all safe (tier-1) items, `n` deselects everything. A live **`N selected (size)`** counter always shows what you have chosen.

**Smart Cleanup** selects only safe (tier-1) items for you and jumps straight to review.

The **Review** screen is the last gate. It lists every item that will be removed, grouped by stage, each tagged `→ Trash (recoverable)` or `permanent`, with per-stage and grand totals. Press `Enter` to proceed; danger items additionally require `y`.

```
  read-only — nothing is deleted until you confirm

  Review — 2.0 GB across 6 items

  Node.js                                   1.6 GB
    ✓ node_modules  ~/work/api              1.2 GB   → Trash (recoverable)
    ✓ node_modules  ~/work/web              380 MB   → Trash (recoverable)
    ✓ npm cache                              41 MB   → Trash (recoverable)

  Docker                                    402 MB
    ✓ build cache                           402 MB   permanent

  System                                     18 MB
    ✓ Homebrew cache                         18 MB   → Trash (recoverable)

  Total → Trash: 1.6 GB      Permanent: 402 MB

  ▸ Enter  clean      y  confirm danger items      Esc  back
```

### Keys

| Key | Action | Key | Action |
|-----|--------|-----|--------|
| `↑ ↓` / `j k` | Navigate | `a` | Select all safe |
| `Enter` | Open / Continue | `n` | Deselect all |
| `Space` | Select / Deselect | `s` | Cycle sort (Explorer) |
| `y` | Confirm danger item | `/` | Search / filter (Explorer) |
| `?` | Help overlay | `Esc` | Back |
| `q` | Quit | | |

## Scriptable CLI

The dashboard is the default for humans; the CLI remains for scripts, CI, and SSH sessions.

| Command | Experience |
|---------|-----------|
| `clean-me` | Interactive dashboard (falls back to `scan` with no TTY) |
| `clean-me scan [--json]` | Read-only report |
| `clean-me clean [--yes] [--stage node] [--tier 2] [--force] [--no-trash]` | Non-interactive cleanup with per-item confirmation |
| `clean-me history` | List past cleanup runs (what was moved to Trash) |
| `clean-me restore` | Restore the most recent run from Trash (undo) |
| `clean-me doctor` | Diagnose detected stages, tools and permissions |
| `clean-me --version` | Print the version |

Further persistent flags apply to every command:

| Flag | Effect |
|------|--------|
| `--profile NAME` | Apply a scope profile: `laptop`, `server`, `docker`, `node`, `python`, or one you define. Profiles change *scope only* (stages, tier, paths, min-size) — never the safety flags. |
| `--min-size 200MB` | Large-files threshold. Accepts `200MB`, `1.5GB`, `200M`, `1G`, `500K` or a plain byte count; an unparseable value is ignored and the stage default is used. |
| `--exclude PATH` | Extra path to protect, repeatable. Merged on top of the `~/.config/clean-me/exclude` allow-list. |

### History & undo

Every cleanup run is logged to `~/.config/clean-me/history/log.jsonl` with each item's original path and Trash location. `clean-me history` lists runs; `clean-me restore` moves the most recent run's items back out of the Trash. In the dashboard, the **History** menu item does the same with one-key restore per run.

### Config & profiles

Drop a `~/.config/clean-me/config.json` to define reusable profiles:

```json
{
  "defaultProfile": "laptop",
  "profiles": {
    "work": { "stages": ["node", "docker", "caches"], "tier": 2, "paths": ["~/code"] }
  }
}
```

Built-in profiles (`laptop`, `server`, `docker`, `node`, `python`) work without any config. Profiles only widen or narrow *what is scanned* — they can never turn off trash-first or auto-approve danger-tier deletions.

### Safety model

- **Read-only by default** — `scan` and `clean --dry-run` never modify anything.
- **Trash-first** — confirmed deletions move to the OS trash; empty it to finalize.
- **Risk tiers** — `safe` (default) → `opt-in` → `danger` (needs `--force`); danger is never auto-selected.
- **Revalidation** — paths are re-checked and symlinks refused immediately before deletion.
- **Project-aware** — a `node_modules` is only offered when its parent is a real project (has `package.json`).
- **Protected paths** — credential and config directories (`.ssh`, `.aws`, `.gnupg`, `.kube`, `.cargo`, `.docker`) are never touched, regardless of tier.
- **Allow-list** — list one path per line in `~/.config/clean-me/exclude` to protect your own paths; `--exclude PATH` adds more on the fly. `clean-me doctor` reports this file's location and whether it exists.
- **Guards** — refuses to run as root without `--allow-root`; refuses to delete in a non-interactive shell without `--yes`.

## Stages

| Stage | What it does | Deletes? |
|-------|--------------|----------|
| **Docker** | Prunes build cache, stopped containers, unused images (tier 2), volumes (tier 3) | yes (permanent — Docker has no trash) |
| **System** | Homebrew / apt / dnf package caches, systemd journal vacuum | yes (permanent) |
| **Node.js** | `node_modules` (trash-first; pnpm/bun link-aware) + npm / yarn / pnpm / bun caches | yes (trash-first) |
| **Python** | `__pycache__`, tool caches, pip cache, virtualenvs (tier 3) | yes (trash-first) |
| **Xcode** | DerivedData, device support, archives and simulator junk | yes (trash-first) |
| **Caches** | Per-app and per-tool `~/Library/Caches` / `~/.cache` entries, sized individually | yes (trash-first) |
| **Trash** | Reports and empties the OS trash to finalize reclaimed space | yes (permanent) |
| **Large Files** | Finds individual files above a size threshold (see `--min-size`) | **report-only** |
| **Git** | Audits merged / upstream-gone / stale (>90d) branches | **report-only** |
| **Redis** | Audits memory usage and biggest keys (read-only, non-blocking) | **report-only** |

Git branches and Redis keys are **audit-only** by design — too easy to break by deleting blindly, so clean-me shows what to review and leaves the decision to you.

## Architecture

The cleanup engine is independent of the UI: stages only *scan* and describe each item's action; one `cleaner` *executes*; the TUI, CLI, and JSON output are interchangeable front ends over the same engine.

```
internal/
  models/    pure data (Item, Tier, Action, ScanResult)
  scanner/   Stage contract + concurrent aggregator
  cleaner/   executor + trash-first deletion
  safety/    tier gating, confirmation, pre-delete revalidation
  stages/    docker · system · node · python · xcode · caches · trash · largefiles · git · redis  (scan-only)
  app/ tui/  Bubble Tea dashboard
```

## Roadmap

Cleanup history & recovery UI, cleanup profiles (laptop / Docker-heavy / production), and scheduled read-only scans. (The disk-usage explorer now ships as the read-only Space Explorer.)

## License

MIT
