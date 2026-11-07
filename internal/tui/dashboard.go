package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

var menuItems = []struct{ title, desc string }{
	{"Smart Cleanup", "Review and clean all safe, reclaimable items."},
	{"Explore Categories", "Docker, Node.js, Python, System, Git, Redis."},
}

// RenderDashboard draws screen 1.
func RenderDashboard(res models.ScanResult, menuCursor int, scanning bool, spinner string, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("clean-me"))
	b.WriteString("\n")
	osName := res.OS
	if osName == "" {
		osName = "local machine"
	}
	b.WriteString(SubtitleStyle.Render("Safe system cleanup · " + osName + " · local machine"))
	b.WriteString("\n\n")

	if scanning {
		b.WriteString(spinner + " Scanning your system…\n")
		b.WriteString("\n" + footer("q quit"))
		return b.String()
	}

	b.WriteString(BigStyle.Render(report.Humanize(res.TotalReclaimable())))
	b.WriteString(SubtitleStyle.Render("  potentially reclaimable"))
	b.WriteString("\n")
	b.WriteString(TitleStyle.Render(fmt.Sprintf("%d", res.OpportunityCount())))
	b.WriteString(SubtitleStyle.Render("  cleanup opportunities"))
	b.WriteString("\n\n")

	for _, c := range res.Categories {
		status := ""
		switch {
		case !c.Available:
			status = SubtitleStyle.Render("not available")
		case c.ReportOnly:
			status = SubtitleStyle.Render("audit only")
		default:
			status = report.Humanize(c.Reclaimable())
		}
		b.WriteString(fmt.Sprintf("  %-10s %s\n", c.Title, status))
	}
	b.WriteString("\n")

	b.WriteString(SubtitleStyle.Render("What would you like to do?") + "\n\n")
	for i, m := range menuItems {
		cursor := "  "
		title := m.title
		if i == 0 {
			title += "  (recommended)"
		}
		if i == menuCursor {
			cursor = SelStyle.Render("▸ ")
			title = SelStyle.Render(title)
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, title))
		b.WriteString("    " + SubtitleStyle.Render(m.desc) + "\n")
	}

	b.WriteString("\n" + footer("↑↓ navigate · enter select · q quit"))
	return b.String()
}
