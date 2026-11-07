// Package git is a report-only stage: it audits local branches that look
// abandoned (merged, upstream gone, or stale) across every repo under the given
// paths. It never deletes — branch removal is out of scope.
package git

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/run"
)

type Stage struct{}

func (Stage) Name() string        { return "git" }
func (Stage) Title() string       { return "Git" }
func (Stage) Description() string { return "Merged, stale and upstream-gone branches" }
func (Stage) ReportOnly() bool    { return true }
func (Stage) Available() bool     { return run.Look("git") }

const (
	maxDepth  = 6
	staleDays = 90
)

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
			if strings.Count(filepath.Clean(p), string(os.PathSeparator))-base > maxDepth {
				return filepath.SkipDir
			}
			if d.Name() == ".git" {
				items = append(items, AuditRepo(ctx, filepath.Dir(p))...)
				return filepath.SkipDir
			}
			return nil
		})
	}
	return items, nil
}

// AuditRepo returns report-only items for abandoned-looking branches in a repo.
func AuditRepo(ctx context.Context, repo string) []models.Item {
	def := defaultBranch(ctx, repo)
	current, _ := run.Capture(ctx, "git", "-C", repo, "symbolic-ref", "--short", "HEAD")
	current = strings.TrimSpace(current)

	merged := map[string]bool{}
	if def != "" {
		out, _ := run.Capture(ctx, "git", "-C", repo, "branch", "--merged", def, "--format", "%(refname:short)")
		for _, b := range strings.Fields(out) {
			merged[b] = true
		}
	}

	out, err := run.Capture(ctx, "git", "-C", repo,
		"for-each-ref", "--format", "%(refname:short)\t%(committerdate:unix)\t%(upstream:track)", "refs/heads/")
	if err != nil {
		return nil
	}

	var items []models.Item
	now := time.Now().Unix()
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cols := strings.SplitN(line, "\t", 3)
		branch := cols[0]
		if branch == def || branch == current {
			continue
		}
		var ts int64
		if len(cols) > 1 {
			ts, _ = strconv.ParseInt(strings.TrimSpace(cols[1]), 10, 64)
		}
		track := ""
		if len(cols) > 2 {
			track = cols[2]
		}
		days := int((now - ts) / 86400)

		var reasons []string
		if merged[branch] {
			reasons = append(reasons, "merged into "+def)
		}
		if strings.Contains(track, "gone") {
			reasons = append(reasons, "upstream gone")
		}
		if days >= staleDays {
			reasons = append(reasons, fmt.Sprintf("stale (%dd old)", days))
		}
		if len(reasons) == 0 {
			continue
		}
		items = append(items, models.Item{
			Stage:   "git",
			Path:    fmt.Sprintf("%s › %s", repo, branch),
			Purpose: strings.Join(reasons, ", ") + " — review & `git branch -d` manually",
			Tier:    models.TierSafe,
			Action:  models.ActionReport,
		})
	}
	return items
}

func defaultBranch(ctx context.Context, repo string) string {
	if out, err := run.Capture(ctx, "git", "-C", repo, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil {
		return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(out), "origin/"))
	}
	for _, name := range []string{"main", "master"} {
		if _, err := run.Capture(ctx, "git", "-C", repo, "rev-parse", "--verify", name); err == nil {
			return name
		}
	}
	return ""
}
