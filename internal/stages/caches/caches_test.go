package caches

import (
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

// SubdirItems emits one item per subdir at/over MinCacheSize, skips smaller
// ones, ".Trash", and clean-me's own cache.
func TestSubdirItems(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "big", "data"), MinCacheSize+10)
	writeFile(t, filepath.Join(root, "small", "data"), 100)
	writeFile(t, filepath.Join(root, ".Trash", "data"), MinCacheSize+10)
	writeFile(t, filepath.Join(root, "clean-me", "data"), MinCacheSize+10)

	items := SubdirItems(root, nil)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d: %+v", len(items), items)
	}
	it := items[0]
	if filepath.Base(it.Path) != "big" {
		t.Errorf("wrong subdir surfaced: %s", it.Path)
	}
	if it.Purpose != "big cache" {
		t.Errorf("purpose = %q", it.Purpose)
	}
	if it.Tier != models.TierSafe || it.Action != models.ActionTrash {
		t.Errorf("tier/action = %v/%v, want safe/trash", it.Tier, it.Action)
	}
}

// SubdirItems respects a user exclude.
func TestSubdirItemsExclude(t *testing.T) {
	root := t.TempDir()
	big := filepath.Join(root, "big")
	writeFile(t, filepath.Join(big, "data"), MinCacheSize+10)

	if got := SubdirItems(root, []string{big}); len(got) != 0 {
		t.Fatalf("expected excluded subdir dropped, got %+v", got)
	}
}
