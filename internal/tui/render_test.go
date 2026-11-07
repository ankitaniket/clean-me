package tui

import (
	"testing"

	"github.com/ankitaniket/clean-me/internal/diskstat"
	"github.com/ankitaniket/clean-me/internal/models"
)

func sampleResult() models.ScanResult {
	return models.ScanResult{
		OS: "testos",
		Categories: []models.Category{
			{Stage: "node", Title: "Node.js", Description: "npm and node_modules", Available: true, Items: []models.Item{
				{Stage: "node", Path: "/a/very/long/path/node_modules", Purpose: "reinstallable", Action: models.ActionTrash, Tier: models.TierSafe, Reclaimable: 7_500_000_000, Warn: "pnpm-linked"},
			}},
			{Stage: "git", Title: "Git", Description: "branches", Available: true, ReportOnly: true, Items: []models.Item{
				{Stage: "git", Path: "repo › old", Purpose: "merged", Action: models.ActionReport, Tier: models.TierSafe},
			}},
		},
	}
}

// Render functions must never panic and must produce output at both a normal
// and a very small terminal size.
func TestRendersNonEmpty(t *testing.T) {
	res := sampleResult()
	sel := map[string]bool{res.Categories[0].Items[0].Key(): true}
	disk := diskstat.Usage{Total: 460_000_000_000, Free: 140_000_000_000, Used: 320_000_000_000}
	expView := func(cat models.Category) ExplorerView {
		var items []models.Item
		items = append(items, cat.Items...)
		return ExplorerView{Cat: cat, Items: items, Cursor: 0, Selected: sel, SortBy: "size"}
	}
	spaceEntries := []SpaceEntry{
		{Name: "Projects", Size: 9_000_000_000, IsDir: true},
		{Name: "big.iso", Size: 3_000_000_000, IsDir: false},
	}
	sizes := [][2]int{{80, 24}, {10, 5}}
	for _, s := range sizes {
		w, h := s[0], s[1]
		exp, _ := RenderExplorer(expView(res.Categories[0]), w, h)
		rep, _ := RenderExplorer(expView(res.Categories[1]), w, h)
		spc, _ := RenderSpace(SpaceView{Dir: "/home/you", Entries: spaceEntries}, w, h)
		spcBusy, _ := RenderSpace(SpaceView{Dir: "/home/you", Busy: true, Spinner: "*"}, w, h)
		for name, out := range map[string]string{
			"welcome":    RenderWelcome(w, h),
			"help":       RenderHelp(w, h),
			"dashboard":  RenderDashboard(res, disk, 0, false, "*", w, h),
			"dashNoDisk": RenderDashboard(res, diskstat.Usage{}, 0, false, "*", w, h),
			"scanning":   RenderDashboard(res, disk, 0, true, "*", w, h),
			"categories": RenderCategories(res, 0, w, h),
			"explorer":   exp,
			"report":     rep,
			"space":      spc,
			"spaceBusy":  spcBusy,
			"review":     RenderReview([]models.Item{res.Categories[0].Items[0]}, w, h),
			"progress": RenderProgress(ProgressView{
				Outcomes:   []models.ItemOutcome{{Item: res.Categories[0].Items[0], Reclaimed: 100, TrashedTo: "/t"}},
				Total:      1,
				Done:       true,
				Spinner:    "*",
				HaveFree:   true,
				FreeBefore: 140_000_000_000,
				FreeAfter:  147_500_000_000,
			}, w, h),
		} {
			if out == "" {
				t.Errorf("%s at %dx%d produced empty output", name, w, h)
			}
		}
	}
}
