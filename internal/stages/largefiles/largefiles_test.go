package largefiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ankitaniket/clean-me/internal/models"
)

func writeFile(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Scan honors the threshold: only files at or above MinFileSize surface, and
// they are TierOptIn with ActionTrash.
func TestScanHonorsThreshold(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "big.bin"), 3000)
	writeFile(t, filepath.Join(root, "small.bin"), 500)

	items, err := Stage{}.Scan(context.Background(), models.Options{
		Paths:       []string{root},
		MinFileSize: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d: %+v", len(items), items)
	}
	got := items[0]
	if filepath.Base(got.Path) != "big.bin" {
		t.Errorf("wrong file surfaced: %s", got.Path)
	}
	if got.Tier != models.TierOptIn {
		t.Errorf("tier = %v, want opt-in", got.Tier)
	}
	if got.Action != models.ActionTrash {
		t.Errorf("action = %v, want trash", got.Action)
	}
}

// Scan skips planted Library and node_modules directories entirely.
func TestScanSkipsExcludedDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Library", "huge.bin"), 3000)
	writeFile(t, filepath.Join(root, "node_modules", "huge.bin"), 3000)
	writeFile(t, filepath.Join(root, ".Trash", "huge.bin"), 3000)
	writeFile(t, filepath.Join(root, "keep.bin"), 3000)

	items, err := Stage{}.Scan(context.Background(), models.Options{
		Paths:       []string{root},
		MinFileSize: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d: %+v", len(items), items)
	}
	if filepath.Base(items[0].Path) != "keep.bin" {
		t.Errorf("wrong file surfaced: %s", items[0].Path)
	}
}

// Scan respects a user exclude path.
func TestScanRespectsExclude(t *testing.T) {
	root := t.TempDir()
	secret := filepath.Join(root, "secret")
	writeFile(t, filepath.Join(secret, "huge.bin"), 3000)
	writeFile(t, filepath.Join(root, "keep.bin"), 3000)

	items, err := Stage{}.Scan(context.Background(), models.Options{
		Paths:       []string{root},
		MinFileSize: 1000,
		Exclude:     []string{secret},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if filepath.Dir(it.Path) == secret {
			t.Errorf("excluded path surfaced: %s", it.Path)
		}
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}
