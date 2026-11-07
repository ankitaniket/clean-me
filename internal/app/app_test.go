package app

import (
	"testing"

	"github.com/ankitaniket/clean-me/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

func cannedModel() Model {
	m := New(models.Options{})
	m.st.scanning = false
	m.st.res = models.ScanResult{
		OS: "testos",
		Categories: []models.Category{
			{Stage: "node", Title: "Node.js", Available: true, Items: []models.Item{
				{Stage: "node", Path: "/p/node_modules", Action: models.ActionTrash, Tier: models.TierSafe, Reclaimable: 100},
				{Stage: "node", Path: "/home/.npm", Action: models.ActionCommand, Tier: models.TierSafe, Reclaimable: 50, Command: "npm cache clean --force"},
			}},
			{Stage: "git", Title: "Git", Available: true, ReportOnly: true, Items: []models.Item{
				{Stage: "git", Path: "repo › old", Action: models.ActionReport, Tier: models.TierSafe},
			}},
		},
	}
	return m
}

func send(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	nm, _ := m.Update(msg)
	return nm.(Model)
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestSmartCleanupSelectsTier1(t *testing.T) {
	m := cannedModel()
	m = send(t, m, key("enter")) // menu item 0 = Smart Cleanup
	if m.st.screen != ScreenReview {
		t.Fatalf("screen = %v, want Review", m.st.screen)
	}
	if len(m.st.reviewItems) != 2 {
		t.Fatalf("smart cleanup selected %d items, want 2 (both safe node items)", len(m.st.reviewItems))
	}
}

func TestExploreSelectAllAndReview(t *testing.T) {
	m := cannedModel()
	m = send(t, m, key("down"))  // menuCursor → Explore
	m = send(t, m, key("enter")) // open Categories
	if m.st.screen != ScreenCategories {
		t.Fatalf("screen = %v, want Categories", m.st.screen)
	}
	m = send(t, m, key("enter")) // open node category
	if m.st.screen != ScreenExplorer || m.st.activeCat != 0 {
		t.Fatalf("explorer not opened on node (screen=%v cat=%d)", m.st.screen, m.st.activeCat)
	}
	m = send(t, m, key("a")) // select all safe
	if len(m.st.selected) != 2 {
		t.Fatalf("select-all picked %d, want 2", len(m.st.selected))
	}
	m = send(t, m, key("enter")) // review
	if m.st.screen != ScreenReview || len(m.st.reviewItems) != 2 {
		t.Fatalf("review screen=%v items=%d", m.st.screen, len(m.st.reviewItems))
	}
}

func TestReportCategoryNotSelectable(t *testing.T) {
	m := cannedModel()
	m.st.screen = ScreenExplorer
	m.st.activeCat = 1 // git (report-only)
	m = send(t, m, key("a"))
	m = send(t, m, key("space"))
	if len(m.st.selected) != 0 {
		t.Errorf("report-only items must not be selectable, got %d", len(m.st.selected))
	}
}

func TestSpaceTogglesSelection(t *testing.T) {
	m := cannedModel()
	m.st.screen = ScreenExplorer
	m.st.activeCat = 0
	m.st.itemCursor = 0
	m = send(t, m, key("space"))
	if len(m.st.selected) != 1 {
		t.Fatalf("space should select 1 item, got %d", len(m.st.selected))
	}
	m = send(t, m, key("space"))
	if len(m.st.selected) != 0 {
		t.Fatalf("space again should deselect, got %d", len(m.st.selected))
	}
}
