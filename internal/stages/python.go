package stages

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
	"github.com/ankitaniket/clean-me/internal/stage"
	"github.com/ankitaniket/clean-me/internal/trash"
)

// Python reclaims space from Python projects: regenerated caches
// (__pycache__, tool caches), the pip download cache, and virtualenvs (danger).
type Python struct{}

func (Python) Name() string { return "python" }

func (Python) Available() bool { return true }

const pyMaxDepth = 8

var pyCacheDirs = map[string]string{
	"__pycache__":   "compiled bytecode, regenerated on import",
	".pytest_cache": "pytest cache, regenerated on next run",
	".mypy_cache":   "mypy cache, regenerated on next run",
	".ruff_cache":   "ruff cache, regenerated on next run",
}

func (Python) Scan(ctx context.Context, o stage.Options) ([]stage.Item, error) {
	roots := o.Paths
	if len(roots) == 0 {
		roots = []string{"."}
	}
	var items []stage.Item
	for _, root := range roots {
		base := strings.Count(filepath.Clean(root), string(os.PathSeparator))
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			name := d.Name()
			if name == ".git" {
				return filepath.SkipDir
			}
			if strings.Count(filepath.Clean(p), string(os.PathSeparator))-base > pyMaxDepth {
				return filepath.SkipDir
			}
			if purpose, ok := pyCacheDirs[name]; ok {
				size, _ := sizeutil.DirSize(p)
				items = append(items, stage.Item{
					Stage: "python", Path: p, Purpose: purpose,
					Size: size, Reclaimable: size, Tier: stage.TierSafe,
				})
				return filepath.SkipDir
			}
			if isVenv(p) {
				size, _ := sizeutil.DirSize(p)
				items = append(items, stage.Item{
					Stage: "python", Path: p,
					Purpose: "virtualenv — recreate with `python -m venv` + reinstall",
					Size:    size, Reclaimable: size, Tier: stage.TierDanger,
					Warn: "tier 3: deletes an environment; ensure requirements are captured",
				})
				return filepath.SkipDir
			}
			return nil
		})
	}

	// pip download cache (global, re-downloadable).
	if run.Look("pip") || run.Look("pip3") {
		if home, err := os.UserHomeDir(); err == nil {
			for _, c := range []string{
				filepath.Join(home, "Library", "Caches", "pip"), // macOS
				filepath.Join(home, ".cache", "pip"),            // Linux
			} {
				if size, _ := sizeutil.DirSize(c); size > 0 {
					items = append(items, stage.Item{
						Stage: "python", Path: c,
						Purpose: "pip download cache, re-downloadable",
						Size:    size, Reclaimable: size, Tier: stage.TierSafe,
						Command: "pip cache purge",
					})
					break
				}
			}
		}
	}
	return items, nil
}

func (Python) Apply(ctx context.Context, items []stage.Item, o stage.Options) (stage.Result, error) {
	res := stage.Result{Stage: "python"}
	for _, it := range items {
		if !gated(it, o) {
			res.Skipped++
			continue
		}
		if it.Command != "" {
			args := strings.Fields(it.Command)
			if _, err := run.Run(ctx, o.DryRun, args[0], args[1:]...); err != nil {
				res.Errors = append(res.Errors, err)
				continue
			}
		} else if !o.DryRun {
			if _, err := trash.Remove(it.Path, o.NoTrash); err != nil {
				res.Errors = append(res.Errors, fmt.Errorf("%s: %w", it.Path, err))
				continue
			}
		}
		res.Removed++
		res.Reclaimed += it.Reclaimable
	}
	return res, nil
}

// isVenv heuristically identifies a virtualenv directory.
func isVenv(p string) bool {
	base := filepath.Base(p)
	if base != "venv" && base != ".venv" && base != "env" {
		return false
	}
	return sizeutil.Exists(filepath.Join(p, "pyvenv.cfg")) ||
		sizeutil.Exists(filepath.Join(p, "bin", "activate"))
}
