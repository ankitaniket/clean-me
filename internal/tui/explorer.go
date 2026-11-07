package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// ExplorerView is the data the explorer render needs. Items is a pre-sorted,
// pre-filtered copy owned by the caller (the ScanResult is never mutated).
type ExplorerView struct {
	Cat       models.Category
	Items     []models.Item
	Cursor    int
	Top       int // current top of the scroll window (may be adjusted)
	Selected  map[string]bool
	SortBy    string // "size" | "name" | "tier"
	Filter    string
	Filtering bool
}

// RenderExplorer draws screen 3: compact one-line rows with checkboxes, a
// scroll window that keeps the focused row in view, sort/filter header, and a
// detail panel for the focused item. It returns the rendered string plus the
// (possibly adjusted) window top so the caller can persist it.
func RenderExplorer(v ExplorerView, w, h int) (string, int) {
	cat := v.Cat
	var b strings.Builder

	if cat.ReportOnly {
		b.WriteString(TitleStyle.Render(cat.Title) + "  " + SubtitleStyle.Render("audit only — clean-me never deletes these") + "\n\n")
		for _, it := range cat.Items {
			b.WriteString("  " + projectLabel(it.Path) + "\n")
		}
		if len(cat.Items) == 0 {
			b.WriteString(SubtitleStyle.Render("  Nothing to report.") + "\n")
		}
		b.WriteString("\n" + footer("esc back · ? help · q quit"))
		return b.String(), 0
	}

	items := v.Items

	// Selection stats (across the whole category, independent of filter).
	var selN int
	var selSize int64
	for _, it := range cat.Items {
		if v.Selected[it.Key()] {
			selN++
			selSize += it.Reclaimable
		}
	}

	head := fmt.Sprintf("%s  %s  ·  %s",
		TitleStyle.Render(cat.Title),
		SubtitleStyle.Render(report.Humanize(cat.Reclaimable())+" available"),
		FocusStyle.Render(fmt.Sprintf("%d selected (%s)", selN, report.Humanize(selSize))))
	b.WriteString(head + "   " + RibbonStyle.Render(ribbon) + "\n")

	meta := SubtitleStyle.Render("sort: " + v.SortBy)
	if v.Filtering {
		meta += "   " + FocusStyle.Render("filter: "+v.Filter+"▌")
	} else if v.Filter != "" {
		meta += "   " + FocusStyle.Render("filter: "+v.Filter)
	}
	b.WriteString(meta + "\n\n")

	// Scroll window: reserve rows for header(3) + detail(4) + footer(2).
	rows := h - 9
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
	if top > len(items)-rows && len(items) > rows {
		top = len(items) - rows
	}
	if len(items) <= rows {
		top = 0
	}
	end := top + rows
	if end > len(items) {
		end = len(items)
	}

	for i := top; i < end; i++ {
		it := items[i]
		box := "[ ]"
		if v.Selected[it.Key()] {
			box = CheckStyle.Render("[x]")
		}
		marker := "  "
		name := projectLabel(it.Path)
		if i == cur {
			marker = FocusStyle.Render("▸ ")
			name = FocusStyle.Render(name)
		}
		tag := ""
		if it.Tier >= models.TierDanger {
			tag = "  " + WarnStyle.Render("danger")
		}
		b.WriteString(fmt.Sprintf("%s%s %8s  %s%s\n", marker, box, report.Humanize(it.Reclaimable), name, tag))
	}
	if len(items) == 0 {
		if v.Filter != "" {
			b.WriteString(SubtitleStyle.Render("  No matches for filter.") + "\n")
		} else {
			b.WriteString(SubtitleStyle.Render("  Nothing to clean here.") + "\n")
		}
	} else {
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf("  %d–%d of %d", top+1, end, len(items))) + "\n")
	}

	// Detail panel for the focused item.
	if cur >= 0 && cur < len(items) {
		it := items[cur]
		b.WriteString("\n" + SubtitleStyle.Render(truncate(it.Path, w-2)) + "\n")
		if it.Detail != "" {
			b.WriteString(SubtitleStyle.Render(it.Detail) + "\n")
		}
		b.WriteString(SubtitleStyle.Render(it.Purpose) + "\n")
		if it.Warn != "" {
			b.WriteString(WarnStyle.Render("⚠ "+it.Warn) + "\n")
		}
	}

	b.WriteString("\n" + footer("space select · a all-safe · n none · s sort · / filter · enter review · esc back · ? help"))
	return b.String(), top
}
