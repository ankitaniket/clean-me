package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// RenderCategories draws screen 2: one card per stage. cursor indexes
// res.Categories directly.
func RenderCategories(res models.ScanResult, cursor, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Explore cleanup categories") + "\n")
	b.WriteString(SubtitleStyle.Render("Select a category to inspect its cleanup opportunities.") + "\n\n")

	for i, c := range res.Categories {
		marker := "  "
		name := c.Title
		if i == cursor {
			marker = SelStyle.Render("▸ ")
			name = SelStyle.Render(name)
		}
		var right string
		switch {
		case !c.Available:
			right = SubtitleStyle.Render("not available")
		case c.ScanErr != "":
			right = WarnStyle.Render("scan error")
		case c.ReportOnly:
			right = SubtitleStyle.Render("audit only")
		default:
			right = report.Humanize(c.Reclaimable())
		}
		b.WriteString(fmt.Sprintf("%s%-12s %s\n", marker, name, right))
		b.WriteString("    " + SubtitleStyle.Render(c.Description) + "\n")
	}

	b.WriteString("\n" + footer("↑↓ navigate · enter open · esc back · q quit"))
	return b.String()
}
