package app

import (
	"github.com/ankitaniket/clean-me/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKey routes a keypress to the active screen and returns the next model
// and command.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	// Global quit (disabled mid-clean to avoid leaving a half-run).
	if key == "ctrl+c" || (key == "q" && !m.st.cleaning) {
		return m, tea.Quit
	}
	if m.st.scanning {
		return m, nil // ignore input while the initial scan runs
	}

	switch m.st.screen {
	case ScreenDashboard:
		return m.keyDashboard(key)
	case ScreenCategories:
		return m.keyCategories(key)
	case ScreenExplorer:
		return m.keyExplorer(key)
	case ScreenReview:
		return m.keyReview(key)
	case ScreenProgress:
		return m.keyProgress(key)
	}
	return m, nil
}

func (m Model) keyDashboard(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.st.menuCursor > 0 {
			m.st.menuCursor--
		}
	case "down", "j":
		if m.st.menuCursor < 1 {
			m.st.menuCursor++
		}
	case "enter":
		if m.st.menuCursor == 0 {
			m.st.reviewItems = m.st.smartItems()
			m.st.reviewFrom = ScreenDashboard
			m.st.screen = ScreenReview
		} else {
			m.st.catCursor = 0
			m.st.screen = ScreenCategories
		}
	}
	return m, nil
}

func (m Model) keyCategories(key string) (tea.Model, tea.Cmd) {
	n := len(m.st.res.Categories)
	switch key {
	case "up", "k":
		if m.st.catCursor > 0 {
			m.st.catCursor--
		}
	case "down", "j":
		if m.st.catCursor < n-1 {
			m.st.catCursor++
		}
	case "enter":
		m.st.activeCat = m.st.catCursor
		m.st.itemCursor = 0
		m.st.screen = ScreenExplorer
	case "esc":
		m.st.screen = ScreenDashboard
	}
	return m, nil
}

func (m Model) keyExplorer(key string) (tea.Model, tea.Cmd) {
	items := m.st.activeItems()
	switch key {
	case "up", "k":
		if m.st.itemCursor > 0 {
			m.st.itemCursor--
		}
	case "down", "j":
		if m.st.itemCursor < len(items)-1 {
			m.st.itemCursor++
		}
	case " ", "space":
		if m.st.itemCursor < len(items) {
			m.st.toggle(items[m.st.itemCursor])
		}
	case "a":
		m.st.selectAllSafe()
	case "n":
		m.st.deselectActive()
	case "enter":
		sel := m.st.collectSelected()
		if len(sel) > 0 {
			m.st.reviewItems = sel
			m.st.reviewFrom = ScreenExplorer
			m.st.screen = ScreenReview
		}
	case "esc":
		m.st.screen = ScreenCategories
	}
	return m, nil
}

func (m Model) keyReview(key string) (tea.Model, tea.Cmd) {
	hasDanger := false
	for _, it := range m.st.reviewItems {
		if it.Tier >= models.TierDanger {
			hasDanger = true
			break
		}
	}
	switch key {
	case "esc":
		m.st.screen = m.st.reviewFrom
	case "enter":
		if !hasDanger {
			return m.beginClean()
		}
	case "y":
		if hasDanger {
			return m.beginClean()
		}
	}
	return m, nil
}

func (m Model) keyProgress(key string) (tea.Model, tea.Cmd) {
	if m.st.cleaning {
		return m, nil
	}
	if key == "enter" || key == "esc" {
		// Reset and re-scan to reflect what was reclaimed.
		m.st.selected = map[string]bool{}
		m.st.reviewItems = nil
		m.st.outcomes = nil
		m.st.menuCursor = 0
		m.st.screen = ScreenDashboard
		m.st.scanning = true
		return m, tea.Batch(m.sp.Tick, m.scanCmd())
	}
	return m, nil
}
