package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/ankitaniket/clean-me/internal/report"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Preview what could be reclaimed (read-only, deletes nothing)",
	Long: `scan runs every selected stage read-only and prints the path, purpose and
reclaimable size of each item, grouped by stage, with totals. It is completely
safe — nothing is ever deleted. This is the shareable "here's what it would
free" view.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		items := scanAll(ctx, options(true))
		if flagJSON {
			return report.JSON(os.Stdout, items)
		}
		if len(items) == 0 {
			cmd.Println("Nothing to reclaim — your system is already tidy.")
			return nil
		}
		report.Text(os.Stdout, items)
		cmd.Println("\nRun `clean-me clean` to act on these (with per-item confirmation).")
		return nil
	},
}
