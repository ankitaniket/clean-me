package stages

import (
	"testing"

	"github.com/ankitaniket/clean-me/internal/stage"
)

func TestParseDockerSize(t *testing.T) {
	cases := map[string]int64{
		"0B":            0,
		"512MB":         512_000_000,
		"1.093GB (40%)": 1_093_000_000,
		"2.5GB":         2_500_000_000,
		"213038015":     213038015, // plain bytes (redis used_memory)
		"100MiB":        100 << 20,
		"garbage":       0,
	}
	for in, want := range cases {
		if got := parseDockerSize(in); got != want {
			t.Errorf("parseDockerSize(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestGated(t *testing.T) {
	cases := []struct {
		name    string
		tier    stage.Tier
		maxTier stage.Tier
		force   bool
		want    bool
	}{
		{"safe default", stage.TierSafe, stage.TierSafe, false, true},
		{"optin excluded at tier1", stage.TierOptIn, stage.TierSafe, false, false},
		{"optin included at tier2", stage.TierOptIn, stage.TierOptIn, false, true},
		{"danger needs force", stage.TierDanger, stage.TierDanger, false, false},
		{"danger with force", stage.TierDanger, stage.TierDanger, true, true},
	}
	for _, c := range cases {
		it := stage.Item{Tier: c.tier}
		o := stage.Options{MaxTier: c.maxTier, Force: c.force}
		if got := gated(it, o); got != c.want {
			t.Errorf("%s: gated = %v, want %v", c.name, got, c.want)
		}
	}
}
