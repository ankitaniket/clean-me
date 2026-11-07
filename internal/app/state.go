// Package app is the Bubble Tea front end: it owns application state, keyboard
// navigation, and the tea.Model that drives the five screens. It calls the
// engine (scanner/cleaner) and the tui render functions; the engine knows
// nothing about it.
package app

import (
	"os"
	"sort"
	"strings"

	"github.com/ankitaniket/clean-me/internal/diskstat"
	"github.com/ankitaniket/clean-me/internal/models"
)

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// Screen identifies the active screen.
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenDashboard
	ScreenCategories
	ScreenExplorer
	ScreenReview
	ScreenProgress
	ScreenSpace
)

// sortMode controls how the explorer orders items.
type sortMode int

const (
	sortSize sortMode = iota
	sortName
	sortTier
)

func (s sortMode) next() sortMode {
	switch s {
	case sortSize:
		return sortName
	case sortName:
		return sortTier
	default:
		return sortSize
	}
}

func (s sortMode) label() string {
	switch s {
	case sortName:
		return "name"
	case sortTier:
		return "tier"
	default:
		return "size"
	}
}

// SpaceEntry is one child of a directory in the Space Explorer.
type SpaceEntry struct {
	Name  string
	Size  int64
	IsDir bool
}

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

	// explorer sort/search/scroll state
	sort        sortMode
	filtering   bool
	filter      string
	explorerTop int

	// help overlay (toggled on any screen)
	showHelp bool

	reviewItems []models.Item
	reviewFrom  Screen

	cleaning   bool
	outcomes   []models.ItemOutcome
	progressCh chan models.ItemOutcome

	// disk usage captured before cleaning, to show free before→after
	freeBefore int64
	haveFree   bool

	// Space Explorer state (read-only drill-down)
	spaceDir   string
	spaceStack []string
	spaceCur   int
	spaceTop   int
	spaceCache map[string][]SpaceEntry
	spaceBusy  bool
	spaceErr   string
}

// diskUsage returns disk usage for the SpaceRoot (or home / "/"). Errors yield
// a zero Usage so the render can degrade gracefully.
func (s *State) diskUsage() diskstat.Usage {
	u, err := diskstat.For(s.spaceStartDir())
	if err != nil {
		return diskstat.Usage{}
	}
	return u
}

// spaceStartDir resolves the starting directory for the Space Explorer and the
// dashboard disk bar: Options.SpaceRoot, else $HOME, else "/".
func (s *State) spaceStartDir() string {
	if s.opts.SpaceRoot != "" {
		return s.opts.SpaceRoot
	}
	if h := homeDir(); h != "" {
		return h
	}
	return "/"
}

// sortedActiveItems returns a sorted + filtered copy of the open category's
// items, never mutating the ScanResult.
func (s *State) sortedActiveItems() []models.Item {
	items := s.activeItems()
	out := make([]models.Item, 0, len(items))
	f := strings.ToLower(strings.TrimSpace(s.filter))
	for _, it := range items {
		if f != "" && !strings.Contains(strings.ToLower(it.Path), f) {
			continue
		}
		out = append(out, it)
	}
	switch s.sort {
	case sortName:
		sort.SliceStable(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	case sortTier:
		sort.SliceStable(out, func(i, j int) bool {
			if out[i].Tier != out[j].Tier {
				return out[i].Tier < out[j].Tier
			}
			return out[i].Reclaimable > out[j].Reclaimable
		})
	default: // sortSize
		sort.SliceStable(out, func(i, j int) bool { return out[i].Reclaimable > out[j].Reclaimable })
	}
	return out
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
