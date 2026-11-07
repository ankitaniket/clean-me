// Package diskstat reports total/free/used bytes for the volume holding a
// path, using syscall.Statfs (no external dependency). Works on macOS and
// Linux.
package diskstat

import "syscall"

// Usage describes a filesystem's capacity.
type Usage struct {
	Total int64
	Free  int64
	Used  int64
}

// For returns the disk usage for the volume containing path. On error it
// returns a zero Usage and the error, so callers can degrade gracefully.
func For(path string) (Usage, error) {
	var s syscall.Statfs_t
	if err := syscall.Statfs(path, &s); err != nil {
		return Usage{}, err
	}
	bsize := int64(s.Bsize)
	total := int64(s.Blocks) * bsize
	free := int64(s.Bavail) * bsize
	return Usage{Total: total, Free: free, Used: total - free}, nil
}
