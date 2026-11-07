// Package sizeutil computes directory sizes used throughout the preview. It
// sums regular-file sizes without following symlinks, which keeps results
// deterministic and testable across platforms.
package sizeutil

import (
	"io/fs"
	"os"
	"path/filepath"
)

// DirSize returns the total bytes of regular files under path. Symlinks are not
// followed. Missing paths return 0 with no error so a stage can size optimistic
// candidates cheaply.
func DirSize(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.Type().IsRegular() {
			info, ierr := d.Info()
			if ierr != nil {
				return nil
			}
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// Exists reports whether a path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
