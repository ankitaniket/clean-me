// Package trash implements trash-first deletion: confirmed removals are moved
// to a recoverable location rather than destroyed, unless the user opts into a
// permanent delete with --no-trash. This is what makes clean-me safe to run on
// professional machines.
package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Dir returns the directory clean-me moves deleted paths into. On macOS this is
// the user's ~/.Trash so the Finder can restore them; elsewhere a dedicated
// clean-me trash under the user's data dir.
func Dir() (string, error) {
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

// Remove moves path to the trash (recoverable). If noTrash is true it deletes
// permanently. Returns the final resting place (empty when permanently deleted).
func Remove(path string, noTrash bool) (dest string, err error) {
	if noTrash {
		return "", os.RemoveAll(path)
	}
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	dest = filepath.Join(dir, filepath.Base(path))
	dest = unique(dest)
	if err := os.Rename(path, dest); err != nil {
		// os.Rename fails across filesystems; the caller may retry with
		// --no-trash, but never silently permanent-delete here.
		return "", fmt.Errorf("move to trash failed (try --no-trash to delete in place): %w", err)
	}
	return dest, nil
}

// unique avoids clobbering an existing entry in the trash by suffixing a
// timestamp and, if needed, a counter.
func unique(p string) string {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return p
	}
	stamp := time.Now().Format("20060102-150405")
	base := p + "." + stamp
	cand := base
	for i := 1; ; i++ {
		if _, err := os.Stat(cand); os.IsNotExist(err) {
			return cand
		}
		cand = fmt.Sprintf("%s-%d", base, i)
	}
}
