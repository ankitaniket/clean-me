// Package largefiles surfaces big regular files worth reviewing. These are real
// user data, so every item is TierOptIn (never auto-selected) and trash-first.
package largefiles

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/safety"
)

type Stage struct{}

func (Stage) Name() string        { return "largefiles" }
func (Stage) Title() string       { return "Large Files" }
func (Stage) Description() string { return "Big files worth reviewing" }
func (Stage) ReportOnly() bool    { return false }
func (Stage) Available() bool     { return true }

// DefaultThreshold is the minimum file size when Options.MinFileSize is unset.
const DefaultThreshold = 200 * 1024 * 1024 // 200 MB

const (
	maxDepth = 8
	maxItems = 200
)

// skipDirs are directory names we never descend into: tool state and caches are
// handled by other stages, and user data under them is noise here.
var skipDirs = map[string]bool{
	"Library":      true,
	"Caches":       true,
	"node_modules": true,
	".venv":        true,
	".Trash":       true,
}

func (Stage) Scan(ctx context.Context, o models.Options) ([]models.Item, error) {
	roots := o.Paths
	if len(roots) == 0 {
		if home, err := os.UserHomeDir(); err == nil {
			roots = []string{home}
		} else {
			return nil, nil
		}
	}

	threshold := o.MinFileSize
	if threshold <= 0 {
		threshold = DefaultThreshold
	}

	var items []models.Item
	truncated := false

	for _, root := range roots {
		if truncated {
			break
		}
		base := strings.Count(filepath.Clean(root), string(os.PathSeparator))
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if len(items) >= maxItems {
				truncated = true
				return filepath.SkipAll
			}
			if d.IsDir() {
				name := d.Name()
				if p != root {
					if skipDirs[name] || safety.IsProtected(p, o.Exclude) {
						return filepath.SkipDir
					}
				}
				if strings.Count(filepath.Clean(p), string(os.PathSeparator))-base > maxDepth {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return nil
			}
			info, ierr := d.Info()
			if ierr != nil || info.Size() < threshold {
				return nil
			}
			if safety.IsProtected(p, o.Exclude) {
				return nil
			}
			items = append(items, models.Item{
				Stage:       "largefiles",
				Path:        p,
				Purpose:     "large file",
				Size:        info.Size(),
				Reclaimable: info.Size(),
				Detail:      "modified " + info.ModTime().Format("2006-01-02"),
				Tier:        models.TierOptIn,
				Action:      models.ActionTrash,
			})
			return nil
		})
	}

	// Largest first — most interesting to review.
	sort.Slice(items, func(i, j int) bool { return items[i].Size > items[j].Size })

	if truncated {
		items = append(items, models.Item{
			Stage:   "largefiles",
			Path:    "",
			Purpose: "more large files not shown",
			Warn:    "large-file scan truncated at 200 items; narrow --path to see more",
			Tier:    models.TierOptIn,
			Action:  models.ActionReport,
		})
	}
	return items, nil
}
