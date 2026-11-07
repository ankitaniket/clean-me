package cleaner

import (
	"context"
	"errors"
	"strings"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/run"
)

// execOne performs a single revalidated item's action and returns its outcome.
func execOne(ctx context.Context, it models.Item, o models.Options) models.ItemOutcome {
	out := models.ItemOutcome{Item: it}
	switch it.Action {
	case models.ActionTrash:
		if o.DryRun {
			out.Reclaimed = it.Reclaimable
			return out
		}
		dest, err := trashRemove(it.Path, o.NoTrash)
		if err != nil {
			out.Err = err
			return out
		}
		out.TrashedTo = dest
		out.Reclaimed = it.Reclaimable
	case models.ActionCommand:
		args := strings.Fields(it.Command)
		if len(args) == 0 {
			out.Err = errors.New("empty command")
			return out
		}
		if _, err := run.Run(ctx, o.DryRun, args[0], args[1:]...); err != nil {
			out.Err = err
			return out
		}
		out.Reclaimed = it.Reclaimable // best-effort estimate; tools don't report exact bytes
	default:
		// ActionReport: audit-only, never executed.
	}
	return out
}
