// Package cmd wires the Cobra command tree. `clean-me` with no subcommand
// launches the interactive dashboard (or, with no TTY, prints the read-only
// scan report so it stays useful in pipes, CI and SSH sessions).
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ankitaniket/clean-me/internal/app"
	"github.com/ankitaniket/clean-me/internal/config"
	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/safety"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// Build metadata, overridden at release time via -ldflags
// "-X .../cmd.Version=... -X .../cmd.Commit=... -X .../cmd.Date=...".
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// versionString composes the human-readable version line.
func versionString() string {
	if Commit == "none" && Date == "unknown" {
		return Version
	}
	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date)
}

var (
	flagStages  []string
	flagPaths   []string
	flagTier    int
	flagJSON    bool
	flagForce   bool
	flagYes     bool
	flagNoTrash bool
	flagAllow   bool
	flagMinSize string
	flagExclude []string
	flagProfile string
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
	pf.StringVar(&flagMinSize, "min-size", "", "large-files threshold, e.g. 200MB/1.5GB (default: stage default)")
	pf.StringSliceVar(&flagExclude, "exclude", nil, "extra paths to protect (repeatable); merged with ~/.config/clean-me/exclude")
	pf.StringVar(&flagProfile, "profile", "", "apply a scope profile (laptop/server/docker/node/python or one from your config)")

	rootCmd.Version = versionString()
	rootCmd.SetVersionTemplate("clean-me {{.Version}}\n")
	rootCmd.AddCommand(scanCmd, cleanCmd, doctorCmd, historyCmd, restoreCmd)
}

// baseOptions builds engine options from the shared flags, then layers a
// profile (if --profile is set) on top — profiles adjust scope only, never the
// Force/NoTrash safety flags.
func baseOptions(dryRun bool) models.Options {
	root, _ := os.UserHomeDir()
	o := models.Options{
		Force:       flagForce,
		Yes:         flagYes,
		NoTrash:     flagNoTrash,
		MaxTier:     models.Tier(flagTier),
		Paths:       flagPaths,
		DryRun:      dryRun,
		MinFileSize: parseSize(flagMinSize),
		Exclude:     append(safety.LoadExclude(), flagExclude...),
		SpaceRoot:   root,
	}
	if flagProfile != "" {
		cfg, _ := config.Load("")
		if p, ok := config.Resolve(cfg, flagProfile); ok {
			if stages := config.Apply(p, &o, parseSize); len(flagStages) == 0 && len(stages) > 0 {
				flagStages = stages
			}
		} else {
			fmt.Fprintf(os.Stderr, "note: unknown profile %q — ignoring\n", flagProfile)
		}
	}
	return o
}

// parseSize converts a human size like "200MB", "1.5GB", "200M", "1G" or a
// plain byte count into bytes. On empty input or any parse error it returns 0,
// which stages interpret as "use the stage default".
func parseSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	u := strings.ToUpper(s)
	mult := float64(1)
	switch {
	case strings.HasSuffix(u, "TB") || strings.HasSuffix(u, "T"):
		mult = 1 << 40
		u = strings.TrimSuffix(strings.TrimSuffix(u, "TB"), "T")
	case strings.HasSuffix(u, "GB") || strings.HasSuffix(u, "G"):
		mult = 1 << 30
		u = strings.TrimSuffix(strings.TrimSuffix(u, "GB"), "G")
	case strings.HasSuffix(u, "MB") || strings.HasSuffix(u, "M"):
		mult = 1 << 20
		u = strings.TrimSuffix(strings.TrimSuffix(u, "MB"), "M")
	case strings.HasSuffix(u, "KB") || strings.HasSuffix(u, "K"):
		mult = 1 << 10
		u = strings.TrimSuffix(strings.TrimSuffix(u, "KB"), "K")
	case strings.HasSuffix(u, "B"):
		u = strings.TrimSuffix(u, "B")
	}
	u = strings.TrimSpace(u)
	v, err := strconv.ParseFloat(u, 64)
	if err != nil || v < 0 {
		return 0
	}
	return int64(v * mult)
}

// refuseRoot guards against running with full privileges unless opted in.
func refuseRoot() error {
	if os.Geteuid() == 0 && !flagAllow {
		return fmt.Errorf("refusing to run as root without --allow-root")
	}
	return nil
}
