package stages

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(cmd.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func TestAuditRepoFlagsMergedBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
	git(t, dir, "checkout", "-q", "-b", "feature")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "work")
	git(t, dir, "checkout", "-q", "main")
	git(t, dir, "merge", "-q", "--no-ff", "-m", "merge", "feature")

	items := auditRepo(context.Background(), dir)

	var found bool
	for _, it := range items {
		if strings.Contains(it.Path, "feature") && strings.Contains(it.Purpose, "merged") {
			found = true
		}
		if !it.ReportOnly {
			t.Errorf("git item %q must be report-only", it.Path)
		}
	}
	if !found {
		t.Fatalf("expected merged branch 'feature' to be flagged; got %+v", items)
	}
}
