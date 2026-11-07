package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ankitaniket/clean-me/internal/confirm"
	"github.com/ankitaniket/clean-me/internal/report"
	"github.com/ankitaniket/clean-me/internal/stage"
)

var cleanDryRun bool

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Reclaim space with per-item confirmation (trash-first)",
	Long: `clean previews every item, then asks before acting on each one (default No).
Confirmed deletions go to the trash first so they stay recoverable; pass
--no-trash to delete permanently. Danger-tier items require --force, and in a
non-interactive shell nothing is deleted unless --yes is given.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := refuseRoot(); err != nil {
			return err
		}
		ctx := context.Background()
		o := options(cleanDryRun)

		items := scanAll(ctx, o)
		if len(items) == 0 {
			cmd.Println("Nothing to reclaim — your system is already tidy.")
			return nil
		}
		report.Text(os.Stdout, items)

		if cleanDryRun {
			cmd.Println("\n[dry-run] The commands above would run after confirmation. Nothing was changed.")
			return nil
		}

		if !confirm.Interactive() && !flagYes {
			return fmt.Errorf("refusing to delete in a non-interactive shell without --yes")
		}

		// Group confirmed items per stage, prompting for each actionable item.
		accepted := map[string][]stage.Item{}
		for _, it := range items {
			if it.ReportOnly {
				continue
			}
			if it.Tier > o.MaxTier {
				fmt.Printf("  · skipping %s (tier %s > --tier %d)\n", it.Path, it.Tier, flagTier)
				continue
			}
			if it.Tier >= stage.TierDanger && !flagForce {
				fmt.Printf("  · skipping %s (danger-tier needs --force)\n", it.Path)
				continue
			}
			prompt := fmt.Sprintf("Remove %s (%s, %s)?", it.Path, it.Purpose, report.Humanize(it.Reclaimable))
			if confirm.Ask(prompt, flagYes) {
				accepted[it.Stage] = append(accepted[it.Stage], it)
			}
		}

		var total int64
		for _, s := range selected() {
			acc := accepted[s.Name()]
			if len(acc) == 0 {
				continue
			}
			res, err := s.Apply(ctx, acc, o)
			if err != nil {
				fmt.Fprintf(os.Stderr, "note: %s apply: %v\n", s.Name(), err)
			}
			for _, e := range res.Errors {
				fmt.Fprintf(os.Stderr, "  ! %v\n", e)
			}
			total += res.Reclaimed
		}

		cmd.Printf("\nDone. Reclaimed ~%s", report.Humanize(total))
		if !flagNoTrash {
			cmd.Print(" (moved to trash — empty it to finalize)")
		}
		cmd.Println(".")
		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "preview the exact commands without deleting")
}
