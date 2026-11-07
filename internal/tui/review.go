package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// RenderReview draws screen 4: the complete list of exactly what will be
// removed, grouped by stage, with a Trash/permanent tag per item.
func RenderReview(items []models.Item, w, h int) string {
	var reclaimable int64
	var permanent, trash, danger int
	for _, it := range items {
		reclaimable += it.Reclaimable
		if it.Permanent {
			permanent++
		} else {
			trash++
		}
		if it.Tier >= models.TierDanger {
			danger++
		}
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("Review — exactly what will be cleaned") + "\n\n")
	b.WriteString(fmt.Sprintf("  %d items · %s · %s · %s\n\n",
		len(items),
		BigStyle.Render(report.Humanize(reclaimable)),
		CheckStyle.Render(fmt.Sprintf("%d → Trash (recoverable)", trash)),
		WarnStyle.Render(fmt.Sprintf("%d permanent", permanent))))

	stage := ""
	for _, it := range items {
		if it.Stage != stage {
			stage = it.Stage
			b.WriteString(SubtitleStyle.Render(strings.ToUpper(stage)) + "\n")
		}
		tag := CheckStyle.Render("→ Trash  ")
		if it.Permanent {
			tag = WarnStyle.Render("permanent")
		}
		b.WriteString(fmt.Sprintf("  %8s  %s  %s\n", report.Humanize(it.Reclaimable), tag, truncate(it.Path, w-26)))
	}
	if len(items) == 0 {
		b.WriteString(SubtitleStyle.Render("  Nothing selected.") + "\n")
	}

	b.WriteString("\n")
	if danger > 0 {
		b.WriteString(WarnStyle.Render(fmt.Sprintf("⚠ %d danger item(s) selected.", danger)) + "\n")
		b.WriteString(footer("y confirm · esc back · q quit"))
	} else {
		b.WriteString(footer("enter confirm · esc back · q quit"))
	}
	return b.String()
}
