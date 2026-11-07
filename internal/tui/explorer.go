package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
)

// RenderExplorer draws screen 3: compact one-line rows with checkboxes, plus a
// detail panel for the focused item (so the list itself stays uncluttered).
func RenderExplorer(cat models.Category, cursor int, selected map[string]bool, w, h int) string {
	var b strings.Builder

	if cat.ReportOnly {
		b.WriteString(TitleStyle.Render(cat.Title) + "  " + SubtitleStyle.Render("audit only — clean-me never deletes these") + "\n\n")
		for _, it := range cat.Items {
			b.WriteString("  " + projectLabel(it.Path) + "\n")
		}
		if len(cat.Items) == 0 {
			b.WriteString(SubtitleStyle.Render("  Nothing to report.") + "\n")
		}
		b.WriteString("\n" + footer("esc back · q quit"))
		return b.String()
	}

	// count selection within this category
	var selN int
	var selSize int64
	for _, it := range cat.Items {
		if selected[it.Key()] {
			selN++
			selSize += it.Reclaimable
		}
	}

	head := fmt.Sprintf("%s  %s  ·  %s",
		TitleStyle.Render(cat.Title),
		SubtitleStyle.Render(report.Humanize(cat.Reclaimable())+" available"),
		FocusStyle.Render(fmt.Sprintf("%d selected (%s)", selN, report.Humanize(selSize))))
	b.WriteString(head + "   " + RibbonStyle.Render(ribbon) + "\n\n")

	for i, it := range cat.Items {
		box := "[ ]"
		if selected[it.Key()] {
			box = CheckStyle.Render("[x]")
		}
		marker := "  "
		name := projectLabel(it.Path)
		if i == cursor {
			marker = FocusStyle.Render("▸ ")
			name = FocusStyle.Render(name)
		}
		tag := ""
		if it.Tier >= models.TierDanger {
			tag = "  " + WarnStyle.Render("danger")
		}
		b.WriteString(fmt.Sprintf("%s%s %8s  %s%s\n", marker, box, report.Humanize(it.Reclaimable), name, tag))
	}
	if len(cat.Items) == 0 {
		b.WriteString(SubtitleStyle.Render("  Nothing to clean here.") + "\n")
	}

	// detail panel for the focused item
	if cursor >= 0 && cursor < len(cat.Items) {
		it := cat.Items[cursor]
		b.WriteString("\n" + SubtitleStyle.Render(truncate(it.Path, w-2)) + "\n")
		b.WriteString(SubtitleStyle.Render(it.Purpose) + "\n")
		if it.Warn != "" {
			b.WriteString(WarnStyle.Render("⚠ "+it.Warn) + "\n")
		}
	}

	b.WriteString("\n" + footer("space select · a all-safe · n none · enter review · esc back"))
	return b.String()
}
