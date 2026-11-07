package cleaner

import (
	"context"
	"fmt"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/safety"
)

// ProgressFunc is called with each item's outcome as it completes, so a front
// end can stream progress. It may be nil.
type ProgressFunc func(models.ItemOutcome)

// Execute reclaims the given items one at a time. Each item is gated by tier/
// force and revalidated immediately before deletion; failures are recorded per
// item and never abort the run.
func Execute(ctx context.Context, items []models.Item, o models.Options, progress ProgressFunc) models.Result {
	var res models.Result
	for _, it := range items {
		if !safety.Allowed(it, o) {
			continue
		}
		// Defense in depth: never delete a protected/excluded filesystem path,
		// even if a stage mistakenly offered one.
		if it.Action == models.ActionTrash && safety.IsProtected(it.Path, o.Exclude) {
			res.Outcomes = append(res.Outcomes, models.ItemOutcome{Item: it, Err: fmt.Errorf("protected path, skipped: %s", it.Path)})
			continue
		}
		var out models.ItemOutcome
		if v, err := safety.Revalidate(it); err != nil {
			out = models.ItemOutcome{Item: it, Err: err}
		} else {
			out = execOne(ctx, v, o)
		}
		res.Outcomes = append(res.Outcomes, out)
		if progress != nil {
			progress(out)
		}
	}
	return res
}
