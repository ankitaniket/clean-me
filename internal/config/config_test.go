package config

import (
	"reflect"
	"testing"

	"github.com/ankitaniket/clean-me/internal/models"
)

func fakeParseSize(s string) int64 {
	if s == "200MB" {
		return 200 << 20
	}
	return 0
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := &Config{
		DefaultProfile: "mine",
		Profiles: map[string]Profile{
			"mine": {
				Stages:  []string{"node", "python"},
				Tier:    2,
				Paths:   []string{"/home/me/code"},
				MinSize: "200MB",
				Exclude: []string{"/home/me/keep"},
			},
		},
	}
	if err := Save(dir, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, want)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir() // no config file written
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load of missing file returned error: %v", err)
	}
	if got == nil {
		t.Fatal("Load returned nil Config")
	}
	if got.DefaultProfile != "" || len(got.Profiles) != 0 {
		t.Fatalf("expected empty Config, got %+v", got)
	}
}

func TestResolveBuiltinFallback(t *testing.T) {
	c := &Config{}
	p, ok := Resolve(c, "server")
	if !ok {
		t.Fatal("expected builtin 'server' profile to be found")
	}
	if p.Tier != int(models.TierSafe) {
		t.Fatalf("server tier = %d, want %d", p.Tier, int(models.TierSafe))
	}
	if len(p.Stages) == 0 {
		t.Fatal("expected server profile to limit stages")
	}

	if _, ok := Resolve(c, "does-not-exist"); ok {
		t.Fatal("unexpected resolve of non-existent profile")
	}
}

func TestResolveUserOverridesBuiltin(t *testing.T) {
	c := &Config{
		Profiles: map[string]Profile{
			"server": {Tier: 3, Stages: []string{"custom"}},
		},
	}
	p, ok := Resolve(c, "server")
	if !ok {
		t.Fatal("expected to resolve user 'server' profile")
	}
	if p.Tier != 3 || !reflect.DeepEqual(p.Stages, []string{"custom"}) {
		t.Fatalf("user profile did not override builtin: %+v", p)
	}
}

func TestApply(t *testing.T) {
	o := &models.Options{
		Force:   false,
		NoTrash: false,
		Exclude: []string{"/pre-existing"},
	}
	p := Profile{
		Stages:  []string{"node", "caches"},
		Tier:    2,
		Paths:   []string{"/scan/here"},
		MinSize: "200MB",
		Exclude: []string{"/extra/protected"},
	}
	stages := Apply(p, o, fakeParseSize)

	if !reflect.DeepEqual(stages, []string{"node", "caches"}) {
		t.Fatalf("returned stages = %v", stages)
	}
	if o.MaxTier != models.Tier(2) {
		t.Fatalf("MaxTier = %d, want 2", o.MaxTier)
	}
	if !reflect.DeepEqual(o.Paths, []string{"/scan/here"}) {
		t.Fatalf("Paths = %v", o.Paths)
	}
	if o.MinFileSize != 200<<20 {
		t.Fatalf("MinFileSize = %d, want %d", o.MinFileSize, 200<<20)
	}
	wantExclude := []string{"/pre-existing", "/extra/protected"}
	if !reflect.DeepEqual(o.Exclude, wantExclude) {
		t.Fatalf("Exclude = %v, want %v", o.Exclude, wantExclude)
	}
	// Core safety must be untouched.
	if o.Force {
		t.Fatal("Apply must not set Force")
	}
	if o.NoTrash {
		t.Fatal("Apply must not set NoTrash")
	}
}

func TestApplyZeroValuesLeaveDefaults(t *testing.T) {
	o := &models.Options{MaxTier: models.TierDanger, Paths: []string{"/keep"}, MinFileSize: 42}
	stages := Apply(Profile{}, o, fakeParseSize)
	if stages != nil {
		t.Fatalf("expected nil stages, got %v", stages)
	}
	if o.MaxTier != models.TierDanger {
		t.Fatalf("MaxTier changed to %d", o.MaxTier)
	}
	if !reflect.DeepEqual(o.Paths, []string{"/keep"}) {
		t.Fatalf("Paths changed to %v", o.Paths)
	}
	if o.MinFileSize != 42 {
		t.Fatalf("MinFileSize changed to %d", o.MinFileSize)
	}
}
