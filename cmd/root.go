// Package cmd wires the Cobra command tree and shared flag handling for
// clean-me.
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ankitaniket/clean-me/internal/stage"
	"github.com/ankitaniket/clean-me/internal/stages"
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
	Short: "Safe, multi-stage system cleanup — confirms before deleting, shows path · purpose · size",
	Long: `clean-me reclaims disk space across Docker, OS caches, Node.js and Python
projects, and audits stale git branches and Redis memory.

It never deletes blindly: every item is previewed with its path, purpose and
reclaimable size, deletions go to the trash first (recoverable), and dangerous
operations require an explicit --force. Safe to run on production servers and
professional laptops.`,
	SilenceUsage: true,
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
	pf.BoolVar(&flagJSON, "json", false, "emit machine-readable JSON")
	pf.BoolVar(&flagForce, "force", false, "permit danger-tier (tier 3) and permanent operations")
	pf.BoolVar(&flagYes, "yes", false, "assume yes (non-interactive / CI)")
	pf.BoolVar(&flagNoTrash, "no-trash", false, "delete permanently instead of moving to trash")
	pf.BoolVar(&flagAllow, "allow-root", false, "allow running as root (refused by default)")

	rootCmd.AddCommand(scanCmd, cleanCmd)
}

// allStages is the full registry in display order.
func allStages() []stage.Stage {
	return []stage.Stage{
		stages.Docker{}, stages.System{}, stages.Node{},
		stages.Python{}, stages.Git{}, stages.Redis{},
	}
}

// selected resolves which stages to run: the --stage filter intersected with
// availability. Unavailable stages are reported to stderr so the user knows why
// something was skipped.
func selected() []stage.Stage {
	want := map[string]bool{}
	for _, s := range flagStages {
		want[s] = true
	}
	var out []stage.Stage
	for _, s := range allStages() {
		if len(want) > 0 && !want[s.Name()] {
			continue
		}
		if !s.Available() {
			if len(want) > 0 {
				fmt.Fprintf(os.Stderr, "note: stage %q unavailable (tool/daemon not found), skipping\n", s.Name())
			}
			continue
		}
		out = append(out, s)
	}
	return out
}

func options(dryRun bool) stage.Options {
	return stage.Options{
		DryRun:  dryRun,
		Force:   flagForce,
		Yes:     flagYes,
		NoTrash: flagNoTrash,
		MaxTier: stage.Tier(flagTier),
		Paths:   flagPaths,
	}
}

// refuseRoot guards against running with full privileges unless opted in.
func refuseRoot() error {
	if os.Geteuid() == 0 && !flagAllow {
		return fmt.Errorf("refusing to run as root without --allow-root")
	}
	return nil
}

// scanAll runs every selected stage's read-only Scan and returns the combined
// items.
func scanAll(ctx context.Context, o stage.Options) []stage.Item {
	var items []stage.Item
	for _, s := range selected() {
		found, err := s.Scan(ctx, o)
		if err != nil {
			fmt.Fprintf(os.Stderr, "note: %s scan: %v\n", s.Name(), err)
			continue
		}
		items = append(items, found...)
	}
	return items
}
