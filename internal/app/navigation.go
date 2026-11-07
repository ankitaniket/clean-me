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

	// Help overlay: `?` toggles on any screen (except while typing a filter);
	// while the overlay is open swallow everything but its own toggle/esc.
	if m.st.showHelp {
		if key == "?" || key == "esc" {
			m.st.showHelp = false
		}
		return m, nil
	}
	if key == "?" && !(m.st.screen == ScreenExplorer && m.st.filtering) {
		m.st.showHelp = true
		return m, nil
	}

	switch m.st.screen {
	case ScreenWelcome:
		return m.keyWelcome(key)
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
	case ScreenSpace:
		return m.keySpace(key)
	}
	return m, nil
}

func (m Model) keyWelcome(key string) (tea.Model, tea.Cmd) {
	if key == "enter" {
		m.st.scanning = true
		m.st.screen = ScreenDashboard
		return m, tea.Batch(m.sp.Tick, m.scanCmd())
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
		if m.st.menuCursor < 2 {
			m.st.menuCursor++
		}
	case "enter":
		switch m.st.menuCursor {
		case 0:
			m.st.reviewItems = m.st.smartItems()
			m.st.reviewFrom = ScreenDashboard
			m.st.screen = ScreenReview
		case 1:
			m.st.catCursor = 0
			m.st.screen = ScreenCategories
		case 2:
			return m.openSpace()
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
		m.st.explorerTop = 0
		m.st.filter = ""
		m.st.filtering = false
		m.st.screen = ScreenExplorer
	case "esc":
		m.st.screen = ScreenDashboard
	}
	return m, nil
}

func (m Model) keyExplorer(key string) (tea.Model, tea.Cmd) {
	// Filter (search) mode captures typing until enter/esc.
	if m.st.filtering {
		switch key {
		case "enter":
			m.st.filtering = false // keep the filter
		case "esc":
			m.st.filtering = false
			m.st.filter = "" // clear the filter
		case "backspace":
			if len(m.st.filter) > 0 {
				r := []rune(m.st.filter)
				m.st.filter = string(r[:len(r)-1])
			}
		default:
			if len(key) == 1 { // a single printable rune
				m.st.filter += key
			}
		}
		m.st.itemCursor = 0
		m.st.explorerTop = 0
		return m, nil
	}

	items := m.st.sortedActiveItems()
	switch key {
	case "up", "k":
		if m.st.itemCursor > 0 {
			m.st.itemCursor--
		}
	case "down", "j":
		if m.st.itemCursor < len(items)-1 {
			m.st.itemCursor++
		}
	case "s":
		m.st.sort = m.st.sort.next()
		m.st.itemCursor = 0
		m.st.explorerTop = 0
	case "/":
		m.st.filtering = true
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
		m.st.filter = ""
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

// openSpace enters the Space Explorer at the start directory.
func (m Model) openSpace() (tea.Model, tea.Cmd) {
	m.st.spaceDir = m.st.spaceStartDir()
	m.st.spaceStack = nil
	m.st.spaceCur = 0
	m.st.spaceTop = 0
	m.st.spaceErr = ""
	m.st.screen = ScreenSpace
	return m.loadSpace(m.st.spaceDir)
}

func (m Model) keySpace(key string) (tea.Model, tea.Cmd) {
	if m.st.spaceBusy {
		return m, nil
	}
	entries := m.st.spaceCache[m.st.spaceDir]
	switch key {
	case "up", "k":
		if m.st.spaceCur > 0 {
			m.st.spaceCur--
		}
	case "down", "j":
		if m.st.spaceCur < len(entries)-1 {
			m.st.spaceCur++
		}
	case "enter", "l", "right":
		if m.st.spaceCur < len(entries) {
			e := entries[m.st.spaceCur]
			if e.IsDir {
				child := filepathJoin(m.st.spaceDir, e.Name)
				m.st.spaceStack = append(m.st.spaceStack, m.st.spaceDir)
				m.st.spaceDir = child
				m.st.spaceCur = 0
				m.st.spaceTop = 0
				m.st.spaceErr = ""
				return m.loadSpace(child)
			}
		}
	case "esc", "h", "left":
		if n := len(m.st.spaceStack); n > 0 {
			m.st.spaceDir = m.st.spaceStack[n-1]
			m.st.spaceStack = m.st.spaceStack[:n-1]
			m.st.spaceCur = 0
			m.st.spaceTop = 0
			m.st.spaceErr = ""
		} else {
			m.st.screen = ScreenDashboard
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
