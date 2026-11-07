// Package cleaner executes selected cleanup items. It is the single place where
// deletion happens: stages only describe opportunities, the cleaner carries
// them out (trash-first, with permanent delete only on request).
package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// trashDir returns the directory deleted paths are moved into. On macOS this is
// ~/.Trash (restorable via Finder); elsewhere a dedicated clean-me trash under
// the user's data dir.
func trashDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, ".Trash"), nil
	}
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "clean-me", "trash"), nil
}

// trashRemove moves path to the trash (recoverable). If noTrash is true it
// deletes permanently. Returns the final resting place (empty when permanently
// deleted).
func trashRemove(path string, noTrash bool) (string, error) {
	if noTrash {
		return "", os.RemoveAll(path)
	}
	dir, err := trashDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	dest := uniqueDest(filepath.Join(dir, filepath.Base(path)))
	if err := os.Rename(path, dest); err != nil {
		return "", fmt.Errorf("move to trash failed (try --no-trash to delete in place): %w", err)
	}
	return dest, nil
}

// uniqueDest avoids clobbering an existing trash entry by suffixing a timestamp
// and, if needed, a counter.
func uniqueDest(p string) string {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return p
	}
	base := p + "." + time.Now().Format("20060102-150405")
	cand := base
	for i := 1; ; i++ {
		if _, err := os.Stat(cand); os.IsNotExist(err) {
			return cand
		}
		cand = fmt.Sprintf("%s-%d", base, i)
	}
}
