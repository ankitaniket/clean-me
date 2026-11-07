package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// RenderReview draws screen 4: the exact impact summary before deletion.
func RenderReview(items []models.Item, w, h int) string {
	var reclaimable int64
	var permanent, trash, dangerCount int
	for _, it := range items {
		reclaimable += it.Reclaimable
		if it.Permanent {
			permanent++
		} else {
			trash++
		}
		if it.Tier >= models.TierDanger {
			dangerCount++
		}
	}

	var body strings.Builder
	body.WriteString(TitleStyle.Render("CLEANUP REVIEW") + "\n\n")
	body.WriteString(fmt.Sprintf("Selected items        %d\n", len(items)))
	body.WriteString(fmt.Sprintf("Estimated reclaimable %s\n", report.Humanize(reclaimable)))
	body.WriteString(fmt.Sprintf("Permanent deletions   %d\n", permanent))
	body.WriteString(fmt.Sprintf("Moving to Trash       %d\n", trash))
	body.WriteString("\n")

	// group by stage
	seen := map[string]bool{}
	for _, it := range items {
		if !seen[it.Stage] {
			seen[it.Stage] = true
			body.WriteString(SubtitleStyle.Render(strings.ToUpper(it.Stage)) + "\n")
		}
	}
	body.WriteString("\n")
	for _, it := range items {
		mark := CheckStyle.Render("✓")
		body.WriteString(fmt.Sprintf("%s %8s  %s\n", mark, report.Humanize(it.Reclaimable), truncate(it.Path, w-20)))
	}

	body.WriteString("\n")
	if trash > 0 && permanent == 0 {
		body.WriteString(CheckStyle.Render("This operation is reversible (items move to Trash).") + "\n")
	} else if permanent > 0 {
		body.WriteString(WarnStyle.Render(fmt.Sprintf("%d item(s) will be permanently deleted (no Trash).", permanent)) + "\n")
	}

	var b strings.Builder
	b.WriteString(BoxStyle.Render(body.String()) + "\n")
	if dangerCount > 0 {
		b.WriteString(WarnStyle.Render(fmt.Sprintf("⚠ %d danger-tier item(s) selected — press y to confirm.", dangerCount)) + "\n")
		b.WriteString(footer("y confirm · esc back · q quit"))
	} else {
		b.WriteString(footer("enter clean · esc back · q quit"))
	}
	return b.String()
}
