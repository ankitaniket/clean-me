package stages

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/stage"
)

// Git is a report-only stage: it audits local branches that look abandoned
// (merged, upstream gone, or stale) across every repo under the given paths.
// It never deletes — branch removal is intentionally out of scope for v1.
type Git struct{}

func (Git) Name() string { return "git" }

func (Git) Available() bool { return run.Look("git") }

const gitMaxDepth = 6
const staleDays = 90

func (Git) Scan(ctx context.Context, o stage.Options) ([]stage.Item, error) {
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
			if strings.Count(filepath.Clean(p), string(os.PathSeparator))-base > gitMaxDepth {
				return filepath.SkipDir
			}
			if d.Name() == ".git" {
				repo := filepath.Dir(p)
				items = append(items, auditRepo(ctx, repo)...)
				return filepath.SkipDir
			}
			return nil
		})
	}
	return items, nil
}

// Apply is a no-op: report-only stages never mutate anything.
func (Git) Apply(ctx context.Context, items []stage.Item, o stage.Options) (stage.Result, error) {
	return stage.Result{Stage: "git", Skipped: len(items)}, nil
}

func auditRepo(ctx context.Context, repo string) []stage.Item {
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

	var items []stage.Item
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
		items = append(items, stage.Item{
			Stage:      "git",
			Path:       fmt.Sprintf("%s › %s", repo, branch),
			Purpose:    strings.Join(reasons, ", ") + " — review & `git branch -d` manually",
			Tier:       stage.TierSafe,
			ReportOnly: true,
		})
	}
	return items
}

// defaultBranch resolves the repo's main branch, preferring origin/HEAD, then
// common names.
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
