package node

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ankitaniket/clean-me/internal/models"
)

func TestLinkedManager(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "plain", "node_modules")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := LinkedManager(plain); got != "" {
		t.Errorf("plain: got %q, want empty", got)
	}
	pnpm := filepath.Join(dir, "pnpm", "node_modules")
	if err := os.MkdirAll(filepath.Join(pnpm, ".pnpm"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := LinkedManager(pnpm); got != "pnpm" {
		t.Errorf("pnpm: got %q, want pnpm", got)
	}
}

// Scan must only offer a node_modules that belongs to a real project
// (sibling package.json), never an arbitrary directory of that name.
func TestScanRequiresProjectRoot(t *testing.T) {
	root := t.TempDir()

	// project with package.json → should be offered
	proj := filepath.Join(root, "proj")
	if err := os.MkdirAll(filepath.Join(proj, "node_modules", "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	// bare node_modules with no package.json → must be skipped
	bare := filepath.Join(root, "bare")
	if err := os.MkdirAll(filepath.Join(bare, "node_modules", "y"), 0o755); err != nil {
		t.Fatal(err)
	}

	items, err := Stage{}.Scan(context.Background(), models.Options{Paths: []string{root}})
	if err != nil {
		t.Fatal(err)
	}
	var offered []string
	for _, it := range items {
		if it.Action == models.ActionTrash {
			offered = append(offered, it.Path)
		}
	}
	if len(offered) != 1 {
		t.Fatalf("expected exactly 1 node_modules offered, got %v", offered)
	}
	if filepath.Dir(offered[0]) != proj {
		t.Errorf("offered wrong node_modules: %s", offered[0])
	}
}
