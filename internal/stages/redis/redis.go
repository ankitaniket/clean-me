// Package redis is a report-only stage. Deleting keys from a live Redis is
// dangerous (keys without a TTL are often load-bearing), so clean-me only
// audits memory usage and the biggest keys with read-only, non-blocking
// commands.
package redis

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ankitaniket/clean-me/internal/models"
	"github.com/ankitaniket/clean-me/internal/run"
)

type Stage struct{}

func (Stage) Name() string        { return "redis" }
func (Stage) Title() string       { return "Redis" }
func (Stage) Description() string { return "Memory usage and largest keys" }
func (Stage) ReportOnly() bool    { return true }

func (Stage) Available() bool {
	if !run.Look("redis-cli") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	out, err := run.Capture(ctx, "redis-cli", "ping")
	return err == nil && strings.TrimSpace(out) == "PONG"
}

var memRe = regexp.MustCompile(`(?m)^used_memory:(\d+)`)

func (Stage) Scan(ctx context.Context, o models.Options) ([]models.Item, error) {
	var items []models.Item

	if info, err := run.Capture(ctx, "redis-cli", "info", "memory"); err == nil {
		var used int64
		if m := memRe.FindStringSubmatch(info); m != nil {
			used, _ = strconv.ParseInt(m[1], 10, 64)
		}
		items = append(items, models.Item{
			Stage: "redis", Path: "redis @ localhost:6379",
			Purpose: "current memory footprint (see --bigkeys / no-TTL audit below)",
			Size:    used, Tier: models.TierSafe, Action: models.ActionReport,
		})
	}

	if out, err := run.Capture(ctx, "redis-cli", "--bigkeys"); err == nil {
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Biggest") && strings.Contains(line, "found") {
				items = append(items, models.Item{
					Stage: "redis", Path: "largest key",
					Purpose: strings.Join(strings.Fields(line), " "),
					Tier:    models.TierSafe, Action: models.ActionReport,
				})
			}
		}
	}
	return items, nil
}
