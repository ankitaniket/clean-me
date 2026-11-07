package tui

import (
	"testing"

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
	sizes := [][2]int{{80, 24}, {10, 5}}
	for _, s := range sizes {
		w, h := s[0], s[1]
		for name, out := range map[string]string{
			"dashboard":  RenderDashboard(res, 0, false, "*", w, h),
			"scanning":   RenderDashboard(res, 0, true, "*", w, h),
			"categories": RenderCategories(res, 0, w, h),
			"explorer":   RenderExplorer(res.Categories[0], 0, sel, w, h),
			"report":     RenderExplorer(res.Categories[1], 0, sel, w, h),
			"review":     RenderReview([]models.Item{res.Categories[0].Items[0]}, w, h),
			"progress":   RenderProgress([]models.ItemOutcome{{Item: res.Categories[0].Items[0], Reclaimed: 100, TrashedTo: "/t"}}, 1, true, "*", w, h),
		} {
			if out == "" {
				t.Errorf("%s at %dx%d produced empty output", name, w, h)
			}
		}
	}
}
