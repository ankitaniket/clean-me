package tui

import "strings"

// RenderWelcome is the first screen: a short, scannable intro to what clean-me
// does and its read-only promise. Nothing is scanned until the user chooses to.
func RenderWelcome(w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("clean-me") + "\n")
	b.WriteString(SubtitleStyle.Render("a calm, read-only disk cleanup") + "\n\n")

	b.WriteString(RibbonStyle.Render("Nothing is ever deleted until you review the full list and confirm.") + "\n")
	b.WriteString(SubtitleStyle.Render("First it only looks — a scan touches nothing.") + "\n\n")

	b.WriteString(SubtitleStyle.Render("IT FINDS") + "\n")
	b.WriteString("  " + SubtitleStyle.Render("Docker images & volumes · package caches (npm, pip, brew)") + "\n")
	b.WriteString("  " + SubtitleStyle.Render("Xcode derived data & simulators · node_modules · large files") + "\n\n")

	b.WriteString(SubtitleStyle.Render("TIERS  ") +
		tierStyle(0).Render("safe") + SubtitleStyle.Render(" · ") +
		tierStyle(1).Render("opt-in") + SubtitleStyle.Render(" · ") +
		tierStyle(2).Render("danger") +
		SubtitleStyle.Render("   (danger is never preselected)") + "\n")
	b.WriteString("  " + CheckStyle.Render("→ Trash") + SubtitleStyle.Render(" is recoverable; a few caches purge ") +
		WarnStyle.Render("permanently") + SubtitleStyle.Render(".") + "\n\n")

	b.WriteString(FocusStyle.Render("▸ Scan my system") + "\n\n")
	b.WriteString(footer("enter scan · ? help · q quit"))
	return b.String()
}
