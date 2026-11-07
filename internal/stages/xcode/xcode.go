// Package xcode reclaims space from Apple developer tooling: Xcode DerivedData
// (build intermediates), archives, device-support symbol caches, and the
// CoreSimulator cache. All candidates are trash-first and macOS-only.
package xcode

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/safety"
	"github.com/ankitaniket/clean-me/internal/sizeutil"
)

type Stage struct{}

func (Stage) Name() string        { return "xcode" }
func (Stage) Title() string       { return "Xcode" }
func (Stage) Description() string { return "Xcode derived data, archives & device support" }
func (Stage) ReportOnly() bool    { return false }

// Available reports darwin with an Xcode developer directory present.
func (Stage) Available() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return available(runtime.GOOS, home)
}

// available is the testable core of Available: true only on darwin with an
// ~/Library/Developer/Xcode directory present.
func available(goos, home string) bool {
	if goos != "darwin" {
		return false
	}
	return sizeutil.Exists(filepath.Join(home, "Library", "Developer", "Xcode"))
}

func (Stage) Scan(ctx context.Context, o models.Options) ([]models.Item, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil
	}
	dev := filepath.Join(home, "Library", "Developer")
	xc := filepath.Join(dev, "Xcode")

	type candidate struct {
		path   string
		detail string
		tier   models.Tier
	}
	cands := []candidate{
		{filepath.Join(xc, "DerivedData"), "build intermediates, rebuilt on next build", models.TierSafe},
		{filepath.Join(xc, "Archives"), "real build archives — review", models.TierOptIn},
		{filepath.Join(xc, "iOS DeviceSupport"), "cached device symbols, re-fetched on connect", models.TierOptIn},
		{filepath.Join(xc, "watchOS DeviceSupport"), "cached device symbols, re-fetched on connect", models.TierOptIn},
		{filepath.Join(xc, "tvOS DeviceSupport"), "cached device symbols, re-fetched on connect", models.TierOptIn},
		{filepath.Join(dev, "CoreSimulator", "Caches"), "simulator caches, regenerated", models.TierSafe},
	}

	var items []models.Item
	for _, c := range cands {
		if !sizeutil.Exists(c.path) || safety.IsProtected(c.path, o.Exclude) {
			continue
		}
		size, _ := sizeutil.DirSize(c.path)
		if size <= 0 {
			continue
		}
		items = append(items, models.Item{
			Stage:       "xcode",
			Path:        c.path,
			Purpose:     filepath.Base(c.path),
			Size:        size,
			Reclaimable: size,
			Detail:      c.detail,
			Tier:        c.tier,
			Action:      models.ActionTrash,
		})
	}
	return items, nil
}
