package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// RenderExplorer draws screen 3: selectable items for one category. cursor
// indexes cat.Items; selected is keyed by item.Key().
func RenderExplorer(cat models.Category, cursor int, selected map[string]bool, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render(cat.Title+" cleanup") + "\n")

	if cat.ReportOnly {
		b.WriteString(SubtitleStyle.Render("Audit only — review these; clean-me will not delete them.") + "\n\n")
		for _, it := range cat.Items {
			b.WriteString(fmt.Sprintf("  %s\n", it.Path))
			b.WriteString("    " + SubtitleStyle.Render(it.Purpose) + "\n")
		}
		if len(cat.Items) == 0 {
			b.WriteString(SubtitleStyle.Render("  Nothing to report.") + "\n")
		}
		b.WriteString("\n" + footer("esc back · q quit"))
		return b.String()
	}

	b.WriteString(SubtitleStyle.Render(report.Humanize(cat.Reclaimable())+" available") + "\n\n")

	pathW := w - 40
	if pathW < 12 {
		pathW = 12
	}
	var selTotal int64
	for i, it := range cat.Items {
		check := "[ ]"
		if selected[it.Key()] {
			check = CheckStyle.Render("[x]")
			selTotal += it.Reclaimable
		}
		badge := tierStyle(it.Tier).Render(it.Tier.String())
		marker := "  "
		if i == cursor {
			marker = SelStyle.Render("▸ ")
		}
		line := fmt.Sprintf("%s%s %8s  %-6s  %s", marker, check, report.Humanize(it.Reclaimable), badge, truncate(it.Path, pathW))
		b.WriteString(line + "\n")
		detail := "      " + SubtitleStyle.Render(it.Purpose)
		if it.Warn != "" {
			detail += "  " + WarnStyle.Render("⚠ "+it.Warn)
		}
		b.WriteString(detail + "\n")
	}
	if len(cat.Items) == 0 {
		b.WriteString(SubtitleStyle.Render("  Nothing to clean here.") + "\n")
	}

	b.WriteString("\n" + BigStyle.Render("Selected: "+report.Humanize(selTotal)) + "\n")
	b.WriteString(footer("space select · a all-safe · n none · enter review · esc back · q quit"))
	return b.String()
}
