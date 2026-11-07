package tui

import (
	"fmt"
	"testing"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestPreview renders each screen in plain text for visual review. Run with:
// go test ./internal/tui -run Preview -v
func TestPreview(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("run with -v to see screen previews")
	}
	lipgloss.SetColorProfile(termenv.Ascii) // strip ANSI for a clean capture

	nodeItems := []models.Item{
		{Stage: "node", Path: "/Users/you/Projects/scene-forge/node_modules", Purpose: "node_modules — reinstallable with npm/pnpm/yarn install", Action: models.ActionTrash, Tier: models.TierSafe, Reclaimable: 711_000_000},
		{Stage: "node", Path: "/Users/you/Projects/realtime-video-agent/web/node_modules", Purpose: "node_modules — reinstallable with npm/pnpm/yarn install", Warn: "pnpm-linked: real reclaim is lower; run `pnpm store prune`", Action: models.ActionTrash, Tier: models.TierSafe, Reclaimable: 673_000_000},
		{Stage: "node", Path: "/Users/you/Projects/corelabsfun/web/node_modules", Purpose: "node_modules — reinstallable with npm/pnpm/yarn install", Action: models.ActionTrash, Tier: models.TierSafe, Reclaimable: 658_000_000},
		{Stage: "node", Path: "/Users/you/.npm/_cacache", Purpose: "npm download cache, re-downloadable", Action: models.ActionCommand, Permanent: true, Tier: models.TierSafe, Reclaimable: 8_200_000},
	}
	res := models.ScanResult{OS: "darwin", Categories: []models.Category{
		{Stage: "docker", Title: "Docker", Available: false},
		{Stage: "system", Title: "System", Available: true, Items: []models.Item{{Stage: "system", Reclaimable: 299_000_000, Action: models.ActionCommand, Permanent: true}}},
		{Stage: "node", Title: "Node.js", Available: true, Items: nodeItems},
		{Stage: "python", Title: "Python", Available: true, Items: []models.Item{{Stage: "python", Reclaimable: 1_500_000_000, Action: models.ActionTrash}}},
		{Stage: "git", Title: "Git", Available: true, ReportOnly: true},
		{Stage: "redis", Title: "Redis", Available: false, ReportOnly: true},
	}}
	sel := map[string]bool{nodeItems[0].Key(): true, nodeItems[2].Key(): true}

	reviewItems := []models.Item{
		nodeItems[0], nodeItems[2],
		{Stage: "system", Path: "Homebrew cache (~/Library/Caches/Homebrew)", Permanent: true, Action: models.ActionCommand, Tier: models.TierSafe, Reclaimable: 299_000_000},
		{Stage: "python", Path: "/Users/you/Projects/gram-mitra-ai-backend/.venv", Action: models.ActionTrash, Tier: models.TierDanger, Reclaimable: 386_900_000},
	}

	w, h := 92, 30
	banner := func(s string) { fmt.Printf("\n========== %s ==========\n", s) }
	banner("WELCOME")
	fmt.Println(RenderWelcome(w, h))
	banner("DASHBOARD")
	fmt.Println(RenderDashboard(res, 0, false, "", w, h))
	banner("EXPLORER (Node.js)")
	fmt.Println(RenderExplorer(res.Categories[2], 1, sel, w, h))
	banner("REVIEW")
	fmt.Println(RenderReview(reviewItems, w, h))
}
