package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// ProgressView is the data the progress/results render needs, including the
// disk-free snapshot captured at clean start for a before→after delta.
type ProgressView struct {
	Outcomes   []models.ItemOutcome
	Total      int
	Done       bool
	Spinner    string
	HaveFree   bool
	FreeBefore int64
	FreeAfter  int64
}

// RenderProgress draws screen 5: live per-item progress and the final results.
func RenderProgress(v ProgressView, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("clean-me · cleaning system") + "\n\n")

	for _, o := range v.Outcomes {
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

	if !v.Done {
		b.WriteString("\n" + v.Spinner + fmt.Sprintf(" cleaning (%d/%d)…\n", len(v.Outcomes), v.Total))
		b.WriteString(footer("q quit"))
		return b.String()
	}

	res := models.Result{Outcomes: v.Outcomes}
	b.WriteString("\n" + strings.Repeat("─", min(w, 50)) + "\n\n")
	b.WriteString(TitleStyle.Render("Cleanup completed!") + "\n\n")
	b.WriteString(fmt.Sprintf("Freed space      %s\n", BigStyle.Render(report.Humanize(res.Freed()))))
	b.WriteString(fmt.Sprintf("Moved to Trash   %d items\n", res.Trashed()))
	b.WriteString(fmt.Sprintf("Failed           %d\n", res.Failed()))

	if v.HaveFree {
		delta := v.FreeAfter - v.FreeBefore
		sign := "+"
		if delta < 0 {
			sign = "-"
			delta = -delta
		}
		b.WriteString(fmt.Sprintf("Disk free        %s → %s  (%s%s)\n",
			report.Humanize(v.FreeBefore),
			BigStyle.Render(report.Humanize(v.FreeAfter)),
			sign, report.Humanize(delta)))
	}

	b.WriteString("\n" + footer("enter dashboard · ? help · q quit"))
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
