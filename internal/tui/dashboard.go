package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ankitaniket/clean-me/internal/diskstat"
	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

var menuItems = []struct{ title, hint string }{
	{"Smart Cleanup", "pick the safe items for me"},
	{"Explore", "choose exactly what to clean"},
	{"Space Explorer", "browse what's using disk (read-only)"},
	{"History", "review past cleanups · restore from Trash"},
}

// diskBar draws a single-line proportional usage bar for a volume.
func diskBar(u diskstat.Usage, width int) string {
	if width < 4 {
		width = 4
	}
	if u.Total <= 0 {
		return SubtitleStyle.Render("[" + strings.Repeat("·", width) + "] disk usage unavailable")
	}
	filled := int(float64(u.Used) / float64(u.Total) * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("#", filled) + strings.Repeat("·", width-filled)
	return fmt.Sprintf("[%s] %s used · %s free of %s",
		bar, report.Humanize(u.Used), report.Humanize(u.Free), report.Humanize(u.Total))
}

// RenderDashboard draws screen 1: a disk bar, the reclaimable total, the
// per-category breakdown (largest first), and a 3-item menu.
func RenderDashboard(res models.ScanResult, disk diskstat.Usage, menuCursor int, scanning bool, spinner string, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("clean-me") + "   " + RibbonStyle.Render(ribbon) + "\n\n")

	if scanning {
		b.WriteString(spinner + " scanning — just looking, nothing is touched…\n\n")
		b.WriteString(footer("q quit"))
		return b.String()
	}

	barW := w - 48
	if barW > 24 {
		barW = 24
	}
	if barW < 6 {
		barW = 6
	}
	b.WriteString("   " + SubtitleStyle.Render(diskBar(disk, barW)) + "\n\n")

	b.WriteString("   " + BigStyle.Render(report.Humanize(res.TotalReclaimable())))
	b.WriteString(SubtitleStyle.Render("  reclaimable") + "   ")
	b.WriteString(TitleStyle.Render(fmt.Sprintf("%d", res.OpportunityCount())))
	b.WriteString(SubtitleStyle.Render(" opportunities") + "\n\n")

	// Sort categories by reclaimable desc; report-only/zero/unavailable last.
	order := make([]int, len(res.Categories))
	for i := range order {
		order[i] = i
	}
	rank := func(c models.Category) int64 {
		if !c.Available || c.ReportOnly {
			return -1
		}
		return c.Reclaimable()
	}
	sort.SliceStable(order, func(i, j int) bool {
		return rank(res.Categories[order[i]]) > rank(res.Categories[order[j]])
	})

	for _, idx := range order {
		c := res.Categories[idx]
		var status string
		switch {
		case !c.Available:
			status = SubtitleStyle.Render("—")
		case c.ReportOnly:
			status = SubtitleStyle.Render("audit only")
		default:
			status = report.Humanize(c.Reclaimable())
		}
		b.WriteString(fmt.Sprintf("   %-9s %s\n", c.Title, status))
	}
	b.WriteString("\n")

	for i, m := range menuItems {
		if i == menuCursor {
			b.WriteString(FocusStyle.Render("▸ "+m.title) + SubtitleStyle.Render("  "+m.hint) + "\n")
		} else {
			b.WriteString("  " + m.title + "\n")
		}
	}

	b.WriteString("\n" + footer("↑↓ move · enter select · ? help · q quit"))
	return b.String()
}
