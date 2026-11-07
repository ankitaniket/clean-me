package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

var menuItems = []struct{ title, hint string }{
	{"Smart Cleanup", "pick the safe items for me"},
	{"Explore", "choose exactly what to clean"},
}

// RenderDashboard draws screen 1: compact overview + a 2-item menu.
func RenderDashboard(res models.ScanResult, menuCursor int, scanning bool, spinner string, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("clean-me") + "   " + RibbonStyle.Render(ribbon) + "\n\n")

	if scanning {
		b.WriteString(spinner + " scanning — just looking, nothing is touched…\n\n")
		b.WriteString(footer("q quit"))
		return b.String()
	}

	b.WriteString("   " + BigStyle.Render(report.Humanize(res.TotalReclaimable())))
	b.WriteString(SubtitleStyle.Render("  reclaimable") + "   ")
	b.WriteString(TitleStyle.Render(fmt.Sprintf("%d", res.OpportunityCount())))
	b.WriteString(SubtitleStyle.Render(" opportunities") + "\n\n")

	for _, c := range res.Categories {
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

	b.WriteString("\n" + footer("↑↓ move · enter select · q quit"))
	return b.String()
}
