package tui

import "strings"

// RenderHelp draws the help overlay shown over any screen when `?` is pressed.
// It explains the keys, the tier meanings, and the trash-vs-permanent and
// report-only semantics that make clean-me safe.
func RenderHelp(w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("Help") + "   " + RibbonStyle.Render(ribbon) + "\n\n")

	b.WriteString(SubtitleStyle.Render("KEYS") + "\n")
	for _, r := range [][2]string{
		{"↑ ↓ / k j", "move"},
		{"enter", "select · open · drill in · confirm"},
		{"esc / h", "back · up a level"},
		{"space", "toggle an item for cleaning"},
		{"a / n", "select all-safe / select none"},
		{"s", "cycle sort: size → name → tier (Explore)"},
		{"/", "filter by path (Explore)"},
		{"?", "toggle this help"},
		{"q", "quit"},
	} {
		b.WriteString("  " + FocusStyle.Render(pad(r[0], 12)) + SubtitleStyle.Render(r[1]) + "\n")
	}

	b.WriteString("\n" + SubtitleStyle.Render("TIERS") + "\n")
	b.WriteString("  " + tierStyle(0).Render(pad("safe", 10)) + SubtitleStyle.Render("reinstallable/regenerable — picked by Smart Cleanup") + "\n")
	b.WriteString("  " + tierStyle(1).Render(pad("opt-in", 10)) + SubtitleStyle.Render("fine to clear but you choose it yourself") + "\n")
	b.WriteString("  " + tierStyle(2).Render(pad("danger", 10)) + SubtitleStyle.Render("never preselected; needs an explicit y to confirm") + "\n")

	b.WriteString("\n" + SubtitleStyle.Render("WHAT HAPPENS TO FILES") + "\n")
	b.WriteString("  " + CheckStyle.Render(pad("→ Trash", 12)) + SubtitleStyle.Render("moved to the system Trash — recoverable") + "\n")
	b.WriteString("  " + WarnStyle.Render(pad("permanent", 12)) + SubtitleStyle.Render("removed directly (e.g. cache purge) — not recoverable") + "\n")
	b.WriteString("  " + SubtitleStyle.Render(pad("audit only", 12)) + SubtitleStyle.Render("report-only — clean-me never touches these") + "\n")

	b.WriteString("\n" + RibbonStyle.Render("Nothing is deleted until you review the full list and confirm.") + "\n")
	b.WriteString("\n" + footer("? / esc close · q quit"))
	return b.String()
}

// pad right-pads s to at least n cells.
func pad(s string, n int) string {
	if len(s) >= n {
		return s + " "
	}
	return s + strings.Repeat(" ", n-len(s))
}
