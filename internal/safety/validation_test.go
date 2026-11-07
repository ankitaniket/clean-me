package safety

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ankitaniket/clean-me/internal/models"
)

func TestRevalidate(t *testing.T) {
	dir := t.TempDir()

	// happy path: a real directory, size refreshed.
	real := filepath.Join(dir, "cache")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(real, "f"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	it := models.Item{Action: models.ActionTrash, Path: real}
	got, err := Revalidate(it)
	if err != nil {
		t.Fatalf("real dir: unexpected error %v", err)
	}
	if got.Reclaimable != 5 {
		t.Errorf("real dir: reclaimable = %d, want 5", got.Reclaimable)
	}

	// vanished path must be refused.
	if _, err := Revalidate(models.Item{Action: models.ActionTrash, Path: filepath.Join(dir, "gone")}); err == nil {
		t.Error("vanished path: expected error, got nil")
	}

	// symlink must be refused.
	link := filepath.Join(dir, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if _, err := Revalidate(models.Item{Action: models.ActionTrash, Path: link}); err == nil {
		t.Error("symlink: expected refusal, got nil")
	}

	// command items pass through untouched.
	cmd := models.Item{Action: models.ActionCommand, Path: "docker images", Command: "docker image prune -f"}
	if _, err := Revalidate(cmd); err != nil {
		t.Errorf("command item: unexpected error %v", err)
	}
}
