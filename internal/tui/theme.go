// Package tui holds the presentation layer: a minimal, lazygit-like set of
// lipgloss styles and pure Render functions. Each Render takes plain data plus
// terminal dimensions and returns a string — it owns no state, so the app
// package can drive it without import cycles.
package tui

import (
	"strings"

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
	RibbonStyle   = lipgloss.NewStyle().Foreground(safe).Bold(true)
	FocusStyle    = lipgloss.NewStyle().Foreground(accent).Bold(true)
)

// ribbon is the persistent reassurance shown on every browsing screen.
const ribbon = "● read-only — nothing is deleted until you confirm"

// shortPath keeps the last n path segments (project-relative), hiding the long
// absolute prefix. The full path is shown in the focused detail panel.
func shortPath(p string, n int) string {
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if len(parts) <= n {
		return p
	}
	return "…/" + strings.Join(parts[len(parts)-n:], "/")
}

// genericDirs are parent folder names that carry no project identity — a
// node_modules living directly under one of these needs the grandparent name
// to be recognizable (e.g. "realtime-video-agent/web/node_modules").
var genericDirs = map[string]bool{
	"web": true, "app": true, "apps": true, "packages": true,
	"src": true, "frontend": true, "backend": true, "server": true,
	"client": true, "lib": true, "dist": true,
}

// projectLabel returns a readable identifier for a path shown in list rows.
// For node_modules it disambiguates by showing the project name: if the
// immediate parent is a generic folder (web, app, src, …) it shows the two
// segments before node_modules (project/generic), otherwise just parent/base.
// For anything else it falls back to the last two segments.
func projectLabel(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return path
	}
	base := parts[len(parts)-1]
	if base == "node_modules" && len(parts) >= 3 {
		parent := parts[len(parts)-2]
		if genericDirs[parent] && len(parts) >= 4 {
			// project/generic/node_modules, e.g. realtime-video-agent/web
			return "…/" + strings.Join(parts[len(parts)-3:], "/")
		}
		// parent/node_modules, e.g. scene-forge/node_modules
		return "…/" + strings.Join(parts[len(parts)-2:], "/")
	}
	// fall back to the last two segments
	return shortPath(path, 2)
}

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
