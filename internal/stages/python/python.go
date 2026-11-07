// Package python reclaims space from Python projects: regenerated caches
// (__pycache__, tool caches), the pip download cache, and virtualenvs (danger).
package python

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
)

type Stage struct{}

func (Stage) Name() string        { return "python" }
func (Stage) Title() string       { return "Python" }
func (Stage) Description() string { return "pip cache, bytecode and environments" }
func (Stage) ReportOnly() bool    { return false }
func (Stage) Available() bool     { return true }

const maxDepth = 8

var cacheDirs = map[string]string{
	"__pycache__":   "compiled bytecode, regenerated on import",
	".pytest_cache": "pytest cache, regenerated on next run",
	".mypy_cache":   "mypy cache, regenerated on next run",
	".ruff_cache":   "ruff cache, regenerated on next run",
}

func (Stage) Scan(ctx context.Context, o models.Options) ([]models.Item, error) {
	roots := o.Paths
	if len(roots) == 0 {
		roots = []string{"."}
	}
	var items []models.Item
	for _, root := range roots {
		base := strings.Count(filepath.Clean(root), string(os.PathSeparator))
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			name := d.Name()
			if name == ".Trash" || name == ".git" {
				return filepath.SkipDir
			}
			if strings.Count(filepath.Clean(p), string(os.PathSeparator))-base > maxDepth {
				return filepath.SkipDir
			}
			if purpose, ok := cacheDirs[name]; ok {
				size, _ := sizeutil.DirSize(p)
				items = append(items, models.Item{
					Stage: "python", Path: p, Purpose: purpose,
					Size: size, Reclaimable: size, Tier: models.TierSafe,
					Action: models.ActionTrash,
				})
				return filepath.SkipDir
			}
			if isVenv(p) {
				size, _ := sizeutil.DirSize(p)
				items = append(items, models.Item{
					Stage: "python", Path: p,
					Purpose: "virtualenv — recreate with `python -m venv` + reinstall",
					Size:    size, Reclaimable: size, Tier: models.TierDanger,
					Action: models.ActionTrash,
					Warn:   "tier 3: deletes an environment; ensure requirements are captured",
				})
				return filepath.SkipDir
			}
			return nil
		})
	}

	if run.Look("pip") || run.Look("pip3") {
		if home, err := os.UserHomeDir(); err == nil {
			for _, c := range []string{
				filepath.Join(home, "Library", "Caches", "pip"), // macOS
				filepath.Join(home, ".cache", "pip"),            // Linux
			} {
				if size, _ := sizeutil.DirSize(c); size > 0 {
					purge := "pip cache purge"
					if !run.Look("pip") && run.Look("pip3") {
						purge = "pip3 cache purge"
					}
					items = append(items, models.Item{
						Stage: "python", Path: c,
						Purpose: "pip download cache, re-downloadable",
						Size:    size, Reclaimable: size, Tier: models.TierSafe,
						Action: models.ActionCommand, Permanent: true, Command: purge,
					})
					break
				}
			}
		}
	}
	return items, nil
}

func isVenv(p string) bool {
	base := filepath.Base(p)
	if base != "venv" && base != ".venv" && base != "env" {
		return false
	}
	return sizeutil.Exists(filepath.Join(p, "pyvenv.cfg")) ||
		sizeutil.Exists(filepath.Join(p, "bin", "activate"))
}
