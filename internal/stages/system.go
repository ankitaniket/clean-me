package stages

import (
	"context"
	"regexp"
	"runtime"
	"strings"

	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
	"github.com/ankitaniket/clean-me/internal/stage"
)

// System reclaims OS-level, re-downloadable caches: package-manager caches
// (brew / apt / dnf) and the systemd journal. All commands are permanent (no
// trash), so the preview sizes come from each tool's own reporting.
type System struct{}

func (System) Name() string { return "system" }

func (System) Available() bool {
	return run.Look("brew") || run.Look("apt-get") || run.Look("dnf") || run.Look("journalctl")
}

var freeRe = regexp.MustCompile(`approximately\s+([0-9.]+[KMGT]?i?B)`)

func (System) Scan(ctx context.Context, o stage.Options) ([]stage.Item, error) {
	var items []stage.Item

	if runtime.GOOS == "darwin" && run.Look("brew") {
		// brew has a real dry-run; use it for an accurate preview.
		out, _ := run.Capture(ctx, "brew", "cleanup", "-n")
		var size int64
		if m := freeRe.FindStringSubmatch(out); m != nil {
			size = parseDockerSize(m[1])
		}
		items = append(items, stage.Item{
			Stage: "system", Path: "Homebrew cache (~/Library/Caches/Homebrew)",
			Purpose: "old formula downloads & versions, re-downloadable",
			Size:    size, Reclaimable: size, Tier: stage.TierSafe,
			Command: "brew cleanup",
		})
	}

	if run.Look("apt-get") {
		size, _ := sizeutil.DirSize("/var/cache/apt/archives")
		items = append(items, stage.Item{
			Stage: "system", Path: "/var/cache/apt/archives",
			Purpose: "apt package cache, re-downloadable",
			Size:    size, Reclaimable: size, Tier: stage.TierSafe,
			Command: "apt-get clean",
			Warn:    "may require root (sudo)",
		})
	}

	if run.Look("dnf") {
		size, _ := sizeutil.DirSize("/var/cache/dnf")
		items = append(items, stage.Item{
			Stage: "system", Path: "/var/cache/dnf",
			Purpose: "dnf metadata & package cache, re-downloadable",
			Size:    size, Reclaimable: size, Tier: stage.TierSafe,
			Command: "dnf clean all",
			Warn:    "may require root (sudo)",
		})
	}

	if run.Look("journalctl") {
		out, _ := run.Capture(ctx, "journalctl", "--disk-usage")
		var size int64
		if m := regexp.MustCompile(`take up ([0-9.]+[KMGT]?i?B)`).FindStringSubmatch(out); m != nil {
			size = parseDockerSize(m[1])
		}
		items = append(items, stage.Item{
			Stage: "system", Path: "systemd journal",
			Purpose: "archived logs older than 7 days (running services unaffected)",
			Size:    size, Reclaimable: size / 2, Tier: stage.TierSafe,
			Command: "journalctl --vacuum-time=7d",
			Warn:    "reclaim depends on log age; may require root (sudo)",
		})
	}

	return items, nil
}

func (System) Apply(ctx context.Context, items []stage.Item, o stage.Options) (stage.Result, error) {
	res := stage.Result{Stage: "system"}
	for _, it := range items {
		if !gated(it, o) {
			res.Skipped++
			continue
		}
		args := strings.Fields(it.Command)
		if _, err := run.Run(ctx, o.DryRun, args[0], args[1:]...); err != nil {
			res.Errors = append(res.Errors, err)
			continue
		}
		res.Removed++
		res.Reclaimed += it.Reclaimable
	}
	return res, nil
}
