// Package history persists a log of every cleanup run and supports restoring
// trashed items (undo). Each run is appended as one JSON line to a JSONL log
// under the user's config directory. It depends only on the standard library
// and internal/models (which itself imports nothing), so it is safe to import
// from any layer.
package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ankitaniket/clean-me/internal/models"
)

// RestoredItem records a single trashed item so it can later be moved back from
// the trash to its original location.
type RestoredItem struct {
	Stage        string `json:"stage"`
	OriginalPath string `json:"originalPath"`
	TrashedTo    string `json:"trashedTo"`
	Reclaimed    int64  `json:"reclaimed"`
}

// Run is one recorded cleanup run: when it happened, how many bytes it freed,
// and the restorable (trashed) items it produced.
type Run struct {
	Time  string         `json:"time"`
	Freed int64          `json:"freed"`
	Items []RestoredItem `json:"items"`
}

const (
	dirPerm  os.FileMode = 0700
	filePerm os.FileMode = 0600
	logName              = "log.jsonl"
)

// Dir returns the history directory, <base>/.config/clean-me/history. When base
// is empty it defaults to the user's home directory. The directory is not
// created here; callers that write create it on demand.
func Dir(base string) string {
	if base == "" {
		if home, err := os.UserHomeDir(); err == nil {
			base = home
		}
	}
	return filepath.Join(base, ".config", "clean-me", "history")
}

// logPath returns the JSONL log path for the given base.
func logPath(base string) string {
	return filepath.Join(Dir(base), logName)
}

// Record builds a Run from the successful outcomes and appends it to the log as
// one JSON line. An outcome contributes to Freed when it succeeded (Err == nil);
// it additionally becomes a restorable item when it was trashed (TrashedTo set).
// If there are no restorable items and nothing was freed, Record writes nothing.
func Record(base string, now time.Time, outcomes []models.ItemOutcome) error {
	run := Run{Time: now.Format(time.RFC3339)}
	for _, o := range outcomes {
		if o.Err != nil {
			continue
		}
		run.Freed += o.Reclaimed
		if o.TrashedTo != "" {
			run.Items = append(run.Items, RestoredItem{
				Stage:        o.Item.Stage,
				OriginalPath: o.Item.Path,
				TrashedTo:    o.TrashedTo,
				Reclaimed:    o.Reclaimed,
			})
		}
	}

	if len(run.Items) == 0 && run.Freed == 0 {
		return nil
	}

	line, err := json.Marshal(run)
	if err != nil {
		return fmt.Errorf("history: marshal run: %w", err)
	}

	dir := Dir(base)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("history: create dir: %w", err)
	}

	f, err := os.OpenFile(logPath(base), os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		return fmt.Errorf("history: open log: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("history: write log: %w", err)
	}
	return nil
}

// List reads the log and returns all runs newest-first. A missing log file
// yields an empty slice and a nil error.
func List(base string) ([]Run, error) {
	f, err := os.Open(logPath(base))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Run{}, nil
		}
		return nil, fmt.Errorf("history: open log: %w", err)
	}
	defer f.Close()

	var runs []Run
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var run Run
		if err := json.Unmarshal(line, &run); err != nil {
			return nil, fmt.Errorf("history: parse log line: %w", err)
		}
		runs = append(runs, run)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("history: read log: %w", err)
	}

	// Reverse into newest-first order.
	for i, j := 0, len(runs)-1; i < j; i, j = i+1, j-1 {
		runs[i], runs[j] = runs[j], runs[i]
	}
	if runs == nil {
		return []Run{}, nil
	}
	return runs, nil
}

// Restore moves item.TrashedTo back to item.OriginalPath. It errors if the
// trashed path is empty or missing, or if the original path already exists (to
// avoid clobbering). The parent directory of OriginalPath is created as needed.
// A cross-device rename is surfaced as an error rather than silently ignored.
func Restore(item RestoredItem) error {
	if item.TrashedTo == "" {
		return errors.New("history: restore: trashed path is empty")
	}
	if item.OriginalPath == "" {
		return errors.New("history: restore: original path is empty")
	}
	if _, err := os.Lstat(item.TrashedTo); err != nil {
		return fmt.Errorf("history: restore: trashed item missing: %w", err)
	}
	if _, err := os.Lstat(item.OriginalPath); err == nil {
		return fmt.Errorf("history: restore: original path already exists: %s", item.OriginalPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("history: restore: stat original path: %w", err)
	}

	if parent := filepath.Dir(item.OriginalPath); parent != "" {
		if err := os.MkdirAll(parent, dirPerm); err != nil {
			return fmt.Errorf("history: restore: create parent dir: %w", err)
		}
	}

	if err := os.Rename(item.TrashedTo, item.OriginalPath); err != nil {
		return fmt.Errorf("history: restore: move %s -> %s: %w", item.TrashedTo, item.OriginalPath, err)
	}
	return nil
}

// RestoreRun restores every item in a run, collecting one error per failed
// item. It returns nil when all items restore successfully.
func RestoreRun(run Run) []error {
	var errs []error
	for _, item := range run.Items {
		if err := Restore(item); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
