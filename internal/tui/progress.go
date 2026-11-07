package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// RenderProgress draws screen 5: live per-item progress and the final results.
func RenderProgress(outcomes []models.ItemOutcome, total int, done bool, spinner string, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("clean-me · cleaning system") + "\n\n")

	for _, o := range outcomes {
		if o.Err != nil {
			b.WriteString(WarnStyle.Render("✗ ") + truncate(o.Item.Path, w-30) + "  " + WarnStyle.Render(o.Err.Error()) + "\n")
			continue
		}
		note := report.Humanize(o.Reclaimed)
		if o.TrashedTo != "" {
			note += " → trash"
		}
		b.WriteString(CheckStyle.Render("✓ ") + truncate(o.Item.Path, w-30) + "  " + SubtitleStyle.Render(note) + "\n")
	}

	if !done {
		b.WriteString("\n" + spinner + fmt.Sprintf(" cleaning (%d/%d)…\n", len(outcomes), total))
		b.WriteString(footer("q quit"))
		return b.String()
	}

	res := models.Result{Outcomes: outcomes}
	b.WriteString("\n" + strings.Repeat("─", min(w, 50)) + "\n\n")
	b.WriteString(TitleStyle.Render("Cleanup completed!") + "\n\n")
	b.WriteString(fmt.Sprintf("Freed space      %s\n", BigStyle.Render(report.Humanize(res.Freed()))))
	b.WriteString(fmt.Sprintf("Moved to Trash   %d items\n", res.Trashed()))
	b.WriteString(fmt.Sprintf("Failed           %d\n", res.Failed()))
	b.WriteString("\n" + footer("enter dashboard · q quit"))
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
