// Package cmd wires the Cobra command tree. `clean-me` with no subcommand
// launches the interactive dashboard (or, with no TTY, prints the read-only
// scan report so it stays useful in pipes, CI and SSH sessions).
package cmd

import (
	"fmt"
	"os"

	"github.com/ankitaniket/clean-me/internal/app"
	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	flagStages  []string
	flagPaths   []string
	flagTier    int
	flagJSON    bool
	flagForce   bool
	flagYes     bool
	flagNoTrash bool
	flagAllow   bool
)

var rootCmd = &cobra.Command{
	Use:   "clean-me",
	Short: "Interactive, safe system cleanup — see it, select it, review it, reclaim it",
	Long: `clean-me is an interactive terminal dashboard for reclaiming disk space across
Docker, OS caches, Node.js and Python projects, with audits for stale git
branches and Redis memory.

Run it with no arguments to launch the dashboard. Every item shows its path,
purpose and reclaimable size; deletions move to the trash first; dangerous
operations require explicit confirmation. The scan/clean/doctor subcommands
provide a scriptable, non-interactive surface.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := refuseRoot(); err != nil {
			return err
		}
		if isatty.IsTerminal(os.Stdout.Fd()) {
			return app.Run(baseOptions(false))
		}
		// Non-interactive: fall back to the safe, read-only report.
		return runScan(cmd)
	},
}

// Execute runs the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringSliceVar(&flagStages, "stage", nil, "limit to these stages (docker,system,node,python,git,redis)")
	pf.StringSliceVar(&flagPaths, "path", []string{"."}, "roots to scan for node/python/git")
	pf.IntVar(&flagTier, "tier", 1, "max risk tier to act on (1 safe, 2 opt-in, 3 danger)")
	pf.BoolVar(&flagJSON, "json", false, "emit machine-readable JSON (scan)")
	pf.BoolVar(&flagForce, "force", false, "permit danger-tier (tier 3) operations")
	pf.BoolVar(&flagYes, "yes", false, "assume yes (non-interactive / CI)")
	pf.BoolVar(&flagNoTrash, "no-trash", false, "delete permanently instead of moving to trash")
	pf.BoolVar(&flagAllow, "allow-root", false, "allow running as root (refused by default)")

	rootCmd.AddCommand(scanCmd, cleanCmd, doctorCmd)
}

// baseOptions builds engine options from the shared flags.
func baseOptions(dryRun bool) models.Options {
	return models.Options{
		Force:   flagForce,
		Yes:     flagYes,
		NoTrash: flagNoTrash,
		MaxTier: models.Tier(flagTier),
		Paths:   flagPaths,
		DryRun:  dryRun,
	}
}

// refuseRoot guards against running with full privileges unless opted in.
func refuseRoot() error {
	if os.Geteuid() == 0 && !flagAllow {
		return fmt.Errorf("refusing to run as root without --allow-root")
	}
	return nil
}
