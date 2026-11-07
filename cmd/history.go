package cmd

import (
	"github.com/ankitaniket/clean-me/internal/history"
	"github.com/ankitaniket/clean-me/internal/report"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show past cleanup runs (what was moved to Trash)",
	RunE: func(cmd *cobra.Command, args []string) error {
		runs, err := history.List("")
		if err != nil {
			return err
		}
		if len(runs) == 0 {
			cmd.Println("No cleanup history yet.")
			return nil
		}
		for i, r := range runs {
			cmd.Printf("%2d. %s   freed %s   %d restorable item(s)\n",
				i+1, r.Time, report.Humanize(r.Freed), len(r.Items))
		}
		cmd.Println("\nRun `clean-me restore` to bring the most recent run back from Trash.")
		return nil
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the most recent cleanup from Trash",
	RunE: func(cmd *cobra.Command, args []string) error {
		runs, err := history.List("")
		if err != nil {
			return err
		}
		if len(runs) == 0 {
			cmd.Println("Nothing to restore — no cleanup history.")
			return nil
		}
		run := runs[0]
		errs := history.RestoreRun(run)
		restored := len(run.Items) - len(errs)
		cmd.Printf("Restored %d of %d item(s) from %s.\n", restored, len(run.Items), run.Time)
		for _, e := range errs {
			cmd.Printf("  ! %v\n", e)
		}
		return nil
	},
}
