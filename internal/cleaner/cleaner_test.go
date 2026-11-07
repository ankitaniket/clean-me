package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ankitaniket/clean-me/internal/models"
)

func TestExecuteTrashAndReport(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "big"), []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}

	items := []models.Item{
		{Stage: "node", Path: target, Action: models.ActionTrash, Tier: models.TierSafe, Reclaimable: 10},
		{Stage: "git", Path: "repo › x", Action: models.ActionReport, Tier: models.TierSafe},
	}
	// NoTrash=true deletes in place so the test does not touch the real trash.
	o := models.Options{MaxTier: models.TierDanger, Force: true, NoTrash: true}

	res := Execute(context.Background(), items, o, nil)

	if len(res.Outcomes) != 1 {
		t.Fatalf("expected 1 outcome (report item skipped), got %d", len(res.Outcomes))
	}
	if res.Failed() != 0 {
		t.Fatalf("unexpected failures: %+v", res.Outcomes)
	}
	if res.Freed() != 10 {
		t.Errorf("Freed = %d, want 10", res.Freed())
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("target should have been removed")
	}
}

func TestExecuteDryRun(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cache")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "f"), []byte("abcd"), 0o644); err != nil {
		t.Fatal(err)
	}
	items := []models.Item{{Stage: "x", Path: target, Action: models.ActionTrash, Tier: models.TierSafe, Reclaimable: 4}}
	o := models.Options{MaxTier: models.TierDanger, Force: true, DryRun: true}

	res := Execute(context.Background(), items, o, nil)
	if res.Freed() != 4 {
		t.Errorf("dry-run Freed estimate = %d, want 4", res.Freed())
	}
	if _, err := os.Stat(target); err != nil {
		t.Errorf("dry-run must not delete the path")
	}
}
