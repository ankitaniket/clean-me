package cmd

import (
	"os"

	"github.com/ankitaniket/clean-me/internal/scanner"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose detected stages, tools and permissions",
	Long:  `doctor reports which cleanup stages are available on this machine and flags permission concerns, so you can tell why a stage was skipped.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("clean-me doctor")
		cmd.Println()
		for _, s := range scanner.All() {
			status := "✓ available"
			if !s.Available() {
				status = "— not detected"
			}
			kind := ""
			if s.ReportOnly() {
				kind = " (report-only)"
			}
			cmd.Printf("  %-9s %s%s\n", s.Name(), status, kind)
		}
		cmd.Println()
		if os.Geteuid() == 0 {
			cmd.Println("⚠ running as root — clean-me refuses destructive actions without --allow-root")
		} else {
			cmd.Println("Running as a normal user (recommended).")
		}
		return nil
	},
}
