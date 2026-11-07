package tui

import "strings"

// RenderWelcome is the first screen: a calm, explicit promise that clean-me is
// read-only until the user confirms. Nothing is scanned until they choose to.
func RenderWelcome(w, h int) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("clean-me") + "\n\n")
	b.WriteString("A calm, read-only disk cleanup.\n\n")
	b.WriteString(RibbonStyle.Render("Nothing is ever deleted until you review the full list and confirm.") + "\n")
	b.WriteString(SubtitleStyle.Render("Right now it will only look at your system.") + "\n\n")
	b.WriteString(FocusStyle.Render("▸ Scan my system") + "\n\n")
	b.WriteString(footer("enter scan · q quit"))
	return b.String()
}
