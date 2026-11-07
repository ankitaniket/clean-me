package cmd

import (
	"context"
	"os"

	"github.com/ankitaniket/clean-me/internal/report"
	"github.com/ankitaniket/clean-me/internal/scanner"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Preview what could be reclaimed (read-only, deletes nothing)",
	Long: `scan runs every selected stage read-only and prints each item's path, purpose
and reclaimable size, grouped by category, with totals. Nothing is ever
deleted. Use --json for machine-readable output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan(cmd)
	},
}

// runScan is shared by `scan` and the non-TTY fallback of the root command.
func runScan(cmd *cobra.Command) error {
	// baseOptions first: a --profile may populate the stage filter.
	o := baseOptions(true)
	stages := scanner.Select(flagStages)
	res := scanner.Scan(context.Background(), stages, o)
	if flagJSON {
		return report.JSON(os.Stdout, res)
	}
	if res.OpportunityCount() == 0 && res.TotalReclaimable() == 0 {
		cmd.Println("Nothing to reclaim — your system is already tidy.")
		return nil
	}
	report.Text(os.Stdout, res)
	cmd.Println("\nRun `clean-me` for the interactive dashboard, or `clean-me clean` to act with confirmation.")
	return nil
}
