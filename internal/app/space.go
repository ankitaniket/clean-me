package app

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ankitaniket/clean-me/internal/run"
	tea "github.com/charmbracelet/bubbletea"
)

// spaceLoadedMsg carries the sized children of a directory back to the model.
type spaceLoadedMsg struct {
	dir     string
	entries []SpaceEntry
	err     string
}

func filepathJoin(dir, name string) string { return filepath.Join(dir, name) }

// loadSpace returns a command that lists and sizes the children of dir (via
// `du -sk`), unless they are already cached. It flips spaceBusy so the View can
// show the spinner while sizing runs.
func (m Model) loadSpace(dir string) (tea.Model, tea.Cmd) {
	if _, ok := m.st.spaceCache[dir]; ok {
		return m, nil // already cached — nothing to do
	}
	m.st.spaceBusy = true
	ctx := m.ctx
	cmd := func() tea.Msg {
		entries, errMsg := listSized(ctx, dir)
		return spaceLoadedMsg{dir: dir, entries: entries, err: errMsg}
	}
	return m, tea.Batch(m.sp.Tick, cmd)
}

// listSized reads the directory's children and sizes each with `du -sk`,
// returning them sorted largest-first. Read-only: it never deletes anything.
func listSized(ctx context.Context, dir string) ([]SpaceEntry, string) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err.Error()
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err.Error()
	}
	out := make([]SpaceEntry, 0, len(names))
	for _, name := range names {
		p := filepath.Join(dir, name)
		fi, err := os.Lstat(p)
		if err != nil {
			continue
		}
		isDir := fi.IsDir()
		var size int64
		if isDir {
			size = duSize(ctx, p)
		} else {
			size = fi.Size()
		}
		out = append(out, SpaceEntry{Name: name, Size: size, IsDir: isDir})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Size > out[j].Size })
	return out, ""
}

// duSize returns the size of a directory in bytes via `du -sk`. On error it
// returns 0 so the entry still shows.
func duSize(ctx context.Context, path string) int64 {
	out, err := run.Capture(ctx, "du", "-sk", path)
	if err != nil {
		return 0
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return 0
	}
	kb, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0
	}
	return kb * 1024
}
