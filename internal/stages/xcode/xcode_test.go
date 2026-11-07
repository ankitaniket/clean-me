package xcode

import (
	"os"
	"path/filepath"
	"testing"
)

// available is false when the Xcode developer directory is absent, and false on
// non-darwin regardless, and true only on darwin with the directory present.
func TestAvailable(t *testing.T) {
	home := t.TempDir()

	if available("darwin", home) {
		t.Error("expected false when Xcode dir absent")
	}
	if available("linux", home) {
		t.Error("expected false on non-darwin")
	}

	if err := os.MkdirAll(filepath.Join(home, "Library", "Developer", "Xcode"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !available("darwin", home) {
		t.Error("expected true on darwin with Xcode dir present")
	}
	if available("linux", home) {
		t.Error("expected false on non-darwin even with Xcode dir present")
	}
}
