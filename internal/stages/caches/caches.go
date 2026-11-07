// Package caches reclaims per-application cache directories (one item per app so
// the user can pick), plus the Go build cache and the Gradle caches. Everything
// here is re-downloadable / regenerated, so it is trash-first and TierSafe.
package caches

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/safety"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
)

type Stage struct{}

func (Stage) Name() string        { return "caches" }
func (Stage) Title() string       { return "Caches" }
func (Stage) Description() string { return "App & tool caches" }
func (Stage) ReportOnly() bool    { return false }
func (Stage) Available() bool     { return true }

// MinCacheSize is the per-subdirectory threshold below which a cache directory
// is not worth surfacing on its own. Exposed so tests can reference it.
const MinCacheSize = 20 * 1024 * 1024 // 20 MB

// cacheRoot returns the platform cache root (~/Library/Caches on darwin,
// ~/.cache elsewhere).
func cacheRoot(home string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Caches")
	}
	return filepath.Join(home, ".cache")
}

// SubdirItems enumerates the top-level subdirectories of root, emitting one
// TierSafe trash item per directory whose size is at least MinCacheSize. It
// skips protected paths, ".Trash", and clean-me's own cache. Factored out so it
// can be unit-tested against a temp directory.
func SubdirItems(root string, exclude []string) []models.Item {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var items []models.Item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == ".Trash" || strings.Contains(name, "clean-me") {
			continue
		}
		p := filepath.Join(root, name)
		if safety.IsProtected(p, exclude) {
			continue
		}
		size, _ := sizeutil.DirSize(p)
		if size < MinCacheSize {
			continue
		}
		items = append(items, models.Item{
			Stage:       "caches",
			Path:        p,
			Purpose:     name + " cache",
			Size:        size,
			Reclaimable: size,
			Detail:      name,
			Tier:        models.TierSafe,
			Action:      models.ActionTrash,
		})
	}
	return items
}

func (Stage) Scan(ctx context.Context, o models.Options) ([]models.Item, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil
	}

	items := SubdirItems(cacheRoot(home), o.Exclude)

	// Go build cache — ask the toolchain, fall back to the default location.
	goCache := ""
	if run.Look("go") {
		if out, err := run.Capture(ctx, "go", "env", "GOCACHE"); err == nil {
			goCache = strings.TrimSpace(out)
		}
	}
	if goCache == "" {
		goCache = filepath.Join(home, "Library", "Caches", "go-build")
	}
	items = appendDir(items, goCache, "Go build cache, regenerated on next build", "go-build", o.Exclude)

	// Gradle caches.
	items = appendDir(items, filepath.Join(home, ".gradle", "caches"),
		"Gradle caches, re-downloadable", "gradle", o.Exclude)

	return items, nil
}

// appendDir adds a single TierSafe trash item for dir when it exists, is not
// protected, has size > 0, and isn't already covered by the subdir sweep.
func appendDir(items []models.Item, dir, purpose, detail string, exclude []string) []models.Item {
	if dir == "" || !sizeutil.Exists(dir) || safety.IsProtected(dir, exclude) {
		return items
	}
	clean := filepath.Clean(dir)
	for _, it := range items {
		if filepath.Clean(it.Path) == clean {
			return items // already surfaced by the per-subdir sweep
		}
	}
	size, _ := sizeutil.DirSize(dir)
	if size <= 0 {
		return items
	}
	return append(items, models.Item{
		Stage:       "caches",
		Path:        dir,
		Purpose:     purpose,
		Size:        size,
		Reclaimable: size,
		Detail:      detail,
		Tier:        models.TierSafe,
		Action:      models.ActionTrash,
	})
}
