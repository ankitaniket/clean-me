package tui

import (
	"fmt"
	"strings"

	"github.com/ankitaniket/clean-me/internal/history"
	"github.com/ankitaniket/clean-me/internal/report"
)

// RenderHistory draws the History screen: past cleanup runs, newest first, with
// one-key restore from Trash for the focused run.
func RenderHistory(runs []history.Run, cursor int, msg string, w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Cleanup history") + "   " + RibbonStyle.Render("restore moves items back from Trash") + "\n\n")

	if len(runs) == 0 {
		b.WriteString(SubtitleStyle.Render("No cleanups yet — nothing to restore.") + "\n")
		b.WriteString("\n" + footer("esc back · q quit"))
		return b.String()
	}

	for i, r := range runs {
		marker := "  "
		line := fmt.Sprintf("%s   freed %s   %d item(s)", r.Time, report.Humanize(r.Freed), len(r.Items))
		if i == cursor {
			marker = FocusStyle.Render("▸ ")
			line = FocusStyle.Render(line)
		}
		b.WriteString(marker + line + "\n")
	}

	if msg != "" {
		b.WriteString("\n" + CheckStyle.Render("✓ "+msg) + "\n")
	}
	b.WriteString("\n" + footer("↑↓ move · r restore this run · esc back · q quit"))
	return b.String()
}
