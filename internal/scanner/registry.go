// Package scanner defines the Stage contract and aggregates every stage's
// read-only scan into a single ScanResult that any front end can render.
package scanner

import (
	"context"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/stages/caches"
	"github.com/ankitaniket/clean-me/internal/stages/docker"
	"github.com/ankitaniket/clean-me/internal/stages/git"
	"github.com/ankitaniket/clean-me/internal/stages/largefiles"
	"github.com/ankitaniket/clean-me/internal/stages/node"
	"github.com/ankitaniket/clean-me/internal/stages/python"
	"github.com/ankitaniket/clean-me/internal/stages/redis"
	"github.com/ankitaniket/clean-me/internal/stages/system"
	"github.com/ankitaniket/clean-me/internal/stages/trash"
	"github.com/ankitaniket/clean-me/internal/stages/xcode"
)

// Stage is one cleanup domain. Stages only scan (read-only) and describe how
// each item would be reclaimed via its Action; the cleaner executes. Stages
// never import this interface — they satisfy it structurally — so there is no
// import cycle.
type Stage interface {
	Name() string
	Title() string
	Description() string
	ReportOnly() bool
	Available() bool
	Scan(ctx context.Context, o models.Options) ([]models.Item, error)
}

// All returns the full registry in display order.
func All() []Stage {
	return []Stage{
		docker.Stage{}, system.Stage{}, node.Stage{},
		python.Stage{}, git.Stage{}, redis.Stage{},
		xcode.Stage{}, caches.Stage{}, trash.Stage{}, largefiles.Stage{},
	}
}

// Select returns the stages whose names are in want (all if want is empty).
func Select(want []string) []Stage {
	if len(want) == 0 {
		return All()
	}
	set := map[string]bool{}
	for _, n := range want {
		set[n] = true
	}
	var out []Stage
	for _, s := range All() {
		if set[s.Name()] {
			out = append(out, s)
		}
	}
	return out
}
