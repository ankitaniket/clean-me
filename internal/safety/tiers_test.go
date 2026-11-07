package safety

import (
	"testing"

	"github.com/ankitaniket/clean-me/internal/models"
)

func TestAllowed(t *testing.T) {
	cases := []struct {
		name    string
		tier    models.Tier
		action  models.Action
		maxTier models.Tier
		force   bool
		want    bool
	}{
		{"safe default", models.TierSafe, models.ActionTrash, models.TierSafe, false, true},
		{"optin excluded at tier1", models.TierOptIn, models.ActionCommand, models.TierSafe, false, false},
		{"optin included at tier2", models.TierOptIn, models.ActionCommand, models.TierOptIn, false, true},
		{"danger needs force", models.TierDanger, models.ActionTrash, models.TierDanger, false, false},
		{"danger with force", models.TierDanger, models.ActionTrash, models.TierDanger, true, true},
		{"report never allowed", models.TierSafe, models.ActionReport, models.TierDanger, true, false},
	}
	for _, c := range cases {
		it := models.Item{Tier: c.tier, Action: c.action}
		o := models.Options{MaxTier: c.maxTier, Force: c.force}
		if got := Allowed(it, o); got != c.want {
			t.Errorf("%s: Allowed = %v, want %v", c.name, got, c.want)
		}
	}
}
