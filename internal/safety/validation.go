package safety

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
)

// Revalidate re-checks a filesystem-backed item immediately before deletion and
// returns a refreshed copy (with an up-to-date size). It refuses anything that
// has become unsafe since the scan: a vanished path, a symlink, or a path that
// resolves outside itself (a symlinked ancestor). Command/report items have no
// path to validate and pass through unchanged.
func Revalidate(it models.Item) (models.Item, error) {
	if it.Action != models.ActionTrash {
		return it, nil
	}

	info, err := os.Lstat(it.Path)
	if err != nil {
		return it, fmt.Errorf("path no longer exists: %s", it.Path)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return it, fmt.Errorf("refusing to delete a symlink: %s", it.Path)
	}

	// Reject if any ancestor is a symlink (prevents cleanup from escaping the
	// intended directory).
	abs, err := filepath.Abs(it.Path)
	if err != nil {
		return it, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return it, fmt.Errorf("cannot resolve path: %s", it.Path)
	}
	if resolved != filepath.Clean(abs) {
		return it, fmt.Errorf("refusing to delete via a symlinked path: %s", it.Path)
	}

	// Refresh the size so the reported reclaim reflects reality at delete time.
	if size, serr := sizeutil.DirSize(it.Path); serr == nil {
		it.Size = size
		it.Reclaimable = size
	}
	return it, nil
}
