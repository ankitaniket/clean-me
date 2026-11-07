// Package system reclaims OS-level, re-downloadable caches: package-manager
// caches (brew / apt / dnf) and the systemd journal. All are permanent (no
// trash), so preview sizes come from each tool's own reporting.
package system

import (
	"context"
	"regexp"
	"runtime"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
	"github.com/ankitaniket/clean-me/internal/stages/docker"
)

type Stage struct{}

func (Stage) Name() string        { return "system" }
func (Stage) Title() string       { return "System" }
func (Stage) Description() string { return "Homebrew and OS package caches" }
func (Stage) ReportOnly() bool    { return false }

func (Stage) Available() bool {
	return run.Look("brew") || run.Look("apt-get") || run.Look("dnf") || run.Look("journalctl")
}

var freeRe = regexp.MustCompile(`approximately\s+([0-9.]+[KMGT]?i?B)`)

func (Stage) Scan(ctx context.Context, o models.Options) ([]models.Item, error) {
	var items []models.Item

	if runtime.GOOS == "darwin" && run.Look("brew") {
		out, _ := run.Capture(ctx, "brew", "cleanup", "-n")
		var size int64
		if m := freeRe.FindStringSubmatch(out); m != nil {
			size = docker.ParseSize(m[1])
		}
		items = append(items, models.Item{
			Stage: "system", Path: "Homebrew cache (~/Library/Caches/Homebrew)",
			Purpose: "old formula downloads & versions, re-downloadable",
			Size:    size, Reclaimable: size, Tier: models.TierSafe,
			Action: models.ActionCommand, Permanent: true, Command: "brew cleanup",
		})
	}

	if run.Look("apt-get") {
		size, _ := sizeutil.DirSize("/var/cache/apt/archives")
		items = append(items, models.Item{
			Stage: "system", Path: "/var/cache/apt/archives",
			Purpose: "apt package cache, re-downloadable",
			Size:    size, Reclaimable: size, Tier: models.TierSafe,
			Action: models.ActionCommand, Permanent: true, Command: "apt-get clean",
			Warn: "may require root (sudo)",
		})
	}

	if run.Look("dnf") {
		size, _ := sizeutil.DirSize("/var/cache/dnf")
		items = append(items, models.Item{
			Stage: "system", Path: "/var/cache/dnf",
			Purpose: "dnf metadata & package cache, re-downloadable",
			Size:    size, Reclaimable: size, Tier: models.TierSafe,
			Action: models.ActionCommand, Permanent: true, Command: "dnf clean all",
			Warn: "may require root (sudo)",
		})
	}

	if run.Look("journalctl") {
		out, _ := run.Capture(ctx, "journalctl", "--disk-usage")
		var size int64
		if m := regexp.MustCompile(`take up ([0-9.]+[KMGT]?i?B)`).FindStringSubmatch(out); m != nil {
			size = docker.ParseSize(m[1])
		}
		items = append(items, models.Item{
			Stage: "system", Path: "systemd journal",
			Purpose: "archived logs older than 7 days (running services unaffected)",
			Size:    size, Reclaimable: size / 2, Tier: models.TierSafe,
			Action: models.ActionCommand, Permanent: true, Command: "journalctl --vacuum-time=7d",
			Warn: "reclaim depends on log age; may require root (sudo)",
		})
	}

	return items, nil
}
