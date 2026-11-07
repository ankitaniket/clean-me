package safety

import (
	"fmt"
	"os"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
)

// Revalidate re-checks a filesystem-backed item immediately before deletion and
// returns a refreshed copy (with an up-to-date size). It refuses anything that
// has become unsafe since the scan: a vanished path, or a symlink (which could
// let a delete escape its intended target). Command/report items have no path
// to validate and pass through unchanged.
func Revalidate(it models.Item) (models.Item, error) {
	if it.Action != models.ActionTrash {
		return it, nil
	}

	info, err := os.Lstat(it.Path)
	if err != nil {
		return it, fmt.Errorf("path no longer exists: %s", it.Path)
	}
	// Never operate on a symlink itself — deleting one is surprising and
	// following one would escape the intended directory.
	if info.Mode()&os.ModeSymlink != 0 {
		return it, fmt.Errorf("refusing to delete a symlink: %s", it.Path)
	}

	// Refresh the size so the reported reclaim reflects reality at delete time.
	if size, serr := sizeutil.DirSize(it.Path); serr == nil {
		it.Size = size
		it.Reclaimable = size
	}
	return it, nil
}
