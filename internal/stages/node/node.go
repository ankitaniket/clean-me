// Package node reclaims space from Node.js projects: node_modules directories
// (trash-first, reinstallable) and the npm download cache.
package node

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
)

type Stage struct{}

func (Stage) Name() string        { return "node" }
func (Stage) Title() string       { return "Node.js" }
func (Stage) Description() string { return "npm, pnpm, bun and node_modules" }
func (Stage) ReportOnly() bool    { return false }
func (Stage) Available() bool     { return true }

const maxDepth = 7

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
			if name == ".git" || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			if strings.Count(filepath.Clean(p), string(os.PathSeparator))-base > maxDepth {
				return filepath.SkipDir
			}
			if name == "node_modules" {
				// Only offer a node_modules that belongs to a real project
				// (parent has a package.json) — never an arbitrary directory.
				if !sizeutil.Exists(filepath.Join(filepath.Dir(p), "package.json")) {
					return filepath.SkipDir
				}
				size, _ := sizeutil.DirSize(p)
				it := models.Item{
					Stage: "node", Path: p,
					Purpose:     "node_modules — reinstallable with `npm/pnpm/yarn install`",
					Size:        size,
					Reclaimable: size,
					Tier:        models.TierSafe,
					Action:      models.ActionTrash,
				}
				if pm := LinkedManager(p); pm != "" {
					it.Warn = fmt.Sprintf("%s-linked: real reclaim is lower; run `%s store prune` to free the shared store", pm, pm)
				}
				items = append(items, it)
				return filepath.SkipDir
			}
			return nil
		})
	}

	if run.Look("npm") {
		if home, err := os.UserHomeDir(); err == nil {
			cache := filepath.Join(home, ".npm", "_cacache")
			if size, _ := sizeutil.DirSize(cache); size > 0 {
				items = append(items, models.Item{
					Stage: "node", Path: cache,
					Purpose: "npm download cache, re-downloadable",
					Size:    size, Reclaimable: size, Tier: models.TierSafe,
					Action: models.ActionCommand, Permanent: true, Command: "npm cache clean --force",
				})
			}
		}
	}
	return items, nil
}

// LinkedManager returns "pnpm" or "bun" when a node_modules is populated via a
// shared store/hard-links, where folder size overstates reclaimable bytes.
func LinkedManager(nodeModules string) string {
	if sizeutil.Exists(filepath.Join(nodeModules, ".pnpm")) {
		return "pnpm"
	}
	if sizeutil.Exists(filepath.Join(nodeModules, ".bun")) || sizeutil.Exists(filepath.Join(nodeModules, ".cache", ".bun")) {
		return "bun"
	}
	return ""
}
