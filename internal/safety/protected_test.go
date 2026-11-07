package safety

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsProtected(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	protected := []string{
		filepath.Join(home, ".ssh"),
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".aws", "credentials"),
		filepath.Join(home, ".gnupg"),
		filepath.Join(home, "proj", ".git", "config"),
		home, // the home root itself
		"/",
	}
	for _, p := range protected {
		if !IsProtected(p, nil) {
			t.Errorf("expected protected: %s", p)
		}
	}

	allowed := []string{
		filepath.Join(home, "Projects", "x", "node_modules"),
		filepath.Join(home, "Library", "Caches", "something"),
	}
	for _, p := range allowed {
		if IsProtected(p, nil) {
			t.Errorf("expected NOT protected: %s", p)
		}
	}

	// user-supplied exclude is honored
	keep := filepath.Join(home, "Projects", "keep")
	if !IsProtected(filepath.Join(keep, "node_modules"), []string{keep}) {
		t.Errorf("exclude entry should protect %s/node_modules", keep)
	}
}
