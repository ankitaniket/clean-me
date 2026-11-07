// Package app is the Bubble Tea front end: it owns application state, keyboard
// navigation, and the tea.Model that drives the five screens. It calls the
// engine (scanner/cleaner) and the tui render functions; the engine knows
// nothing about it.
package app

import (
	"github.com/ankitaniket/clean-me/internal/models"
)

// Screen identifies the active screen.
type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenCategories
	ScreenExplorer
	ScreenReview
	ScreenProgress
)

// State is the mutable application state, shared by pointer so async commands
// can observe the latest values.
type State struct {
	screen        Screen
	width, height int

	scanning bool
	res      models.ScanResult
	opts     models.Options

	menuCursor int
	catCursor  int
	activeCat  int
	itemCursor int
	selected   map[string]bool

	reviewItems []models.Item
	reviewFrom  Screen

	cleaning   bool
	outcomes   []models.ItemOutcome
	progressCh chan models.ItemOutcome
}

// activeItems returns the items of the currently open category.
func (s *State) activeItems() []models.Item {
	if s.activeCat < 0 || s.activeCat >= len(s.res.Categories) {
		return nil
	}
	return s.res.Categories[s.activeCat].Items
}

// toggle flips selection of an actionable item.
func (s *State) toggle(it models.Item) {
	if it.ReportOnly() {
		return
	}
	k := it.Key()
	if s.selected[k] {
		delete(s.selected, k)
	} else {
		s.selected[k] = true
	}
}

// selectAllSafe selects every actionable, non-danger item in the open category.
// Danger items are never auto-selected.
func (s *State) selectAllSafe() {
	for _, it := range s.activeItems() {
		if it.ReportOnly() || it.Tier >= models.TierDanger {
			continue
		}
		s.selected[it.Key()] = true
	}
}

// deselectActive clears selection for items in the open category.
func (s *State) deselectActive() {
	for _, it := range s.activeItems() {
		delete(s.selected, it.Key())
	}
}

// collectSelected gathers every selected item across all categories.
func (s *State) collectSelected() []models.Item {
	var out []models.Item
	for _, c := range s.res.Categories {
		for _, it := range c.Items {
			if s.selected[it.Key()] {
				out = append(out, it)
			}
		}
	}
	return out
}

// smartItems returns all available, non-report, tier-1 items (Smart Cleanup).
func (s *State) smartItems() []models.Item {
	var out []models.Item
	for _, c := range s.res.Categories {
		if !c.Available || c.ReportOnly {
			continue
		}
		for _, it := range c.Items {
			if !it.ReportOnly() && it.Tier == models.TierSafe {
				out = append(out, it)
			}
		}
	}
	return out
}
