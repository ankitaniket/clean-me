package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/report"
)

// SpaceEntry is one child of a directory in the Space Explorer. The app builds
// these from its own state (tui must not import app, to avoid a cycle).
type SpaceEntry struct {
	Name  string
	Size  int64
	IsDir bool
}

// SpaceView is the data the Space Explorer render needs.
type SpaceView struct {
	Dir     string
	Entries []SpaceEntry
	Cursor  int
	Top     int
	AtRoot  bool
	Busy    bool
	Err     string
	Spinner string
}

// RenderSpace draws the ncdu-style read-only drill-down: current path, then each
// child with its size, a proportional bar, and a percentage, largest first.
// Returns the rendered string plus the adjusted scroll-window top.
func RenderSpace(v SpaceView, w, h int) (string, int) {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Space Explorer") + "   " + RibbonStyle.Render("● read-only — browse only, nothing is deleted") + "\n")
	b.WriteString(SubtitleStyle.Render(truncate(v.Dir, w-2)) + "\n\n")

	if v.Busy {
		b.WriteString(v.Spinner + " sizing — reading disk usage…\n\n")
		b.WriteString(footer("q quit"))
		return b.String(), v.Top
	}
	if v.Err != "" {
		b.WriteString(WarnStyle.Render("Cannot read this directory: "+v.Err) + "\n\n")
		b.WriteString(footer("esc back · q quit"))
		return b.String(), v.Top
	}

	entries := v.Entries
	var total int64
	for _, e := range entries {
		total += e.Size
	}
	if total <= 0 {
		total = 1
	}

	rows := h - 7
	if rows < 3 {
		rows = 3
	}
	top := v.Top
	cur := v.Cursor
	if cur < top {
		top = cur
	}
	if cur >= top+rows {
		top = cur - rows + 1
	}
	if top < 0 {
		top = 0
	}
	if len(entries) <= rows {
		top = 0
	} else if top > len(entries)-rows {
		top = len(entries) - rows
	}
	end := top + rows
	if end > len(entries) {
		end = len(entries)
	}

	barW := w - 44
	if barW > 20 {
		barW = 20
	}
	if barW < 6 {
		barW = 6
	}

	for i := top; i < end; i++ {
		e := entries[i]
		pct := float64(e.Size) / float64(total)
		filled := int(pct * float64(barW))
		if filled > barW {
			filled = barW
		}
		bar := strings.Repeat("#", filled) + strings.Repeat("·", barW-filled)
		marker := "  "
		name := e.Name
		if e.IsDir {
			name += "/"
		}
		if i == cur {
			marker = FocusStyle.Render("▸ ")
			name = FocusStyle.Render(name)
		}
		b.WriteString(fmt.Sprintf("%s%9s  [%s] %4.0f%%  %s\n",
			marker, report.Humanize(e.Size), bar, pct*100, name))
	}
	if len(entries) == 0 {
		b.WriteString(SubtitleStyle.Render("  (empty)") + "\n")
	} else {
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf("  %d–%d of %d", top+1, end, len(entries))) + "\n")
	}

	back := "esc up"
	if v.AtRoot {
		back = "esc dashboard"
	}
	b.WriteString("\n" + footer("↑↓ move · enter open dir · "+back+" · ? help · q quit"))
	return b.String(), top
}
