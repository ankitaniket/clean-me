package stages

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinkedManager(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "plain", "node_modules")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := linkedManager(plain); got != "" {
		t.Errorf("plain node_modules: got %q, want empty", got)
	}

	pnpm := filepath.Join(dir, "pnpm", "node_modules")
	if err := os.MkdirAll(filepath.Join(pnpm, ".pnpm"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := linkedManager(pnpm); got != "pnpm" {
		t.Errorf("pnpm node_modules: got %q, want pnpm", got)
	}
}
