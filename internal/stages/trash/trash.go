// Package trash reports the system Trash as a single item that empties it. This
// is permanent (not recoverable) so it is TierOptIn and runs via a command.
package trash

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
)

type Stage struct{}

func (Stage) Name() string        { return "trash" }
func (Stage) Title() string       { return "Trash" }
func (Stage) Description() string { return "Empty the system Trash" }
func (Stage) ReportOnly() bool    { return false }

func trashDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".Trash")
}

// Available reports darwin/linux with a ~/.Trash directory present.
func (Stage) Available() bool {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return false
	}
	d := trashDir()
	return d != "" && sizeutil.Exists(d)
}

func (Stage) Scan(ctx context.Context, o models.Options) ([]models.Item, error) {
	d := trashDir()
	if d == "" {
		return nil, nil
	}
	size, _ := sizeutil.DirSize(d)
	if size <= 0 {
		return nil, nil
	}
	return []models.Item{{
		Stage:       "trash",
		Path:        d,
		Purpose:     "system Trash contents",
		Size:        size,
		Reclaimable: size,
		Warn:        "permanently empties the Trash (not recoverable)",
		Tier:        models.TierOptIn,
		Action:      models.ActionCommand,
		Command:     "find " + d + " -mindepth 1 -delete",
		Permanent:   true,
	}}, nil
}
