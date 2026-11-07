package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// RenderCategories draws screen 2: one compact line per stage. cursor indexes
// res.Categories directly.
func RenderCategories(res models.ScanResult, cursor, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Categories") + "   " + RibbonStyle.Render(ribbon) + "\n\n")

	for i, c := range res.Categories {
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
		if i == cursor {
			b.WriteString(FocusStyle.Render(fmt.Sprintf("▸ %-9s", c.Title)) + " " + right + "\n")
		} else {
			b.WriteString(fmt.Sprintf("  %-9s %s\n", c.Title, right))
		}
	}

	b.WriteString("\n" + footer("↑↓ move · enter open · esc back · q quit"))
	return b.String()
}
