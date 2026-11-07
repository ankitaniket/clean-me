// Package tui holds the presentation layer: a minimal, lazygit-like set of
// lipgloss styles and pure Render functions. Each Render takes plain data plus
// terminal dimensions and returns a string — it owns no state, so the app
// package can drive it without import cycles.
package tui

import (
	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/charmbracelet/lipgloss"
)

var (
	accent = lipgloss.Color("45")  // cyan
	dim    = lipgloss.Color("244") // grey
	safe   = lipgloss.Color("42")  // green
	optin  = lipgloss.Color("214") // orange
	danger = lipgloss.Color("203") // red

	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(accent)
	SubtitleStyle = lipgloss.NewStyle().Foreground(dim)
	BigStyle      = lipgloss.NewStyle().Bold(true).Foreground(optin)
	FooterStyle   = lipgloss.NewStyle().Foreground(dim)
	SelStyle      = lipgloss.NewStyle().Bold(true).Foreground(accent)
	WarnStyle     = lipgloss.NewStyle().Foreground(danger)
	BoxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).Padding(0, 1)
	CheckStyle    = lipgloss.NewStyle().Foreground(safe)
)

func tierStyle(t models.Tier) lipgloss.Style {
	switch t {
	case models.TierSafe:
		return lipgloss.NewStyle().Foreground(safe)
	case models.TierOptIn:
		return lipgloss.NewStyle().Foreground(optin)
	case models.TierDanger:
		return lipgloss.NewStyle().Foreground(danger)
	default:
		return lipgloss.NewStyle().Foreground(dim)
	}
}

// truncate shortens s to at most w display cells, adding an ellipsis. For paths
// it keeps the tail (most specific part) visible.
func truncate(s string, w int) string {
	if w <= 1 || lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	if w <= 1 {
		return "…"
	}
	// keep the tail
	keep := w - 1
	if keep >= len(r) {
		return s
	}
	return "…" + string(r[len(r)-keep:])
}

func footer(keys string) string {
	return FooterStyle.Render(keys)
}
