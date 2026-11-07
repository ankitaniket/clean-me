package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ankitaniket/clean-me/internal/cleaner"
	"github.com/ankitaniket/clean-me/internal/history"
	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/report"
	"github.com/ankitaniket/clean-me/internal/safety"
	"github.com/ankitaniket/clean-me/internal/scanner"
	"github.com/spf13/cobra"
)

var cleanDryRun bool

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Reclaim space with per-item confirmation (trash-first)",
	Long: `clean previews every item, then asks before acting on each one (default No).
Confirmed deletions move to the trash first; pass --no-trash to delete
permanently. Danger-tier items require --force, and in a non-interactive shell
nothing is deleted unless --yes is given. For a guided experience, run
` + "`clean-me`" + ` with no arguments instead.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := refuseRoot(); err != nil {
			return err
		}
		ctx := context.Background()
		o := baseOptions(cleanDryRun)

		stages := scanner.Select(flagStages)
		res := scanner.Scan(ctx, stages, o)
		if res.OpportunityCount() == 0 {
			cmd.Println("Nothing to reclaim — your system is already tidy.")
			return nil
		}
		report.Text(os.Stdout, res)

		if cleanDryRun {
			cmd.Println("\n[dry-run] The actions above would run after confirmation. Nothing was changed.")
			return nil
		}
		if !safety.Interactive() && !flagYes {
			return fmt.Errorf("refusing to delete in a non-interactive shell without --yes")
		}

		var accepted []models.Item
		for _, c := range res.Categories {
			for _, it := range c.Items {
				if it.ReportOnly() {
					continue
				}
				if !safety.Allowed(it, o) {
					fmt.Printf("  · skipping %s (tier %s needs --tier/--force)\n", it.Path, it.Tier)
					continue
				}
				prompt := fmt.Sprintf("Remove %s (%s, %s)?", it.Path, it.Purpose, report.Humanize(it.Reclaimable))
				if safety.Ask(prompt, flagYes) {
					accepted = append(accepted, it)
				}
			}
		}

		result := cleaner.Execute(ctx, accepted, o, nil)
		_ = history.Record("", time.Now(), result.Outcomes)
		for _, out := range result.Outcomes {
			if out.Err != nil {
				fmt.Fprintf(os.Stderr, "  ! %s: %v\n", out.Item.Path, out.Err)
			}
		}
		cmd.Printf("\nDone. Reclaimed ~%s", report.Humanize(result.Freed()))
		if result.Trashed() > 0 {
			cmd.Printf(" (%d moved to trash — empty it to finalize)", result.Trashed())
		}
		cmd.Println(".")
		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "preview the exact actions without deleting")
}
