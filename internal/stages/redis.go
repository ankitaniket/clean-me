package stages

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/stage"
)

// Redis is a report-only stage. Deleting keys from a live Redis is dangerous
// (keys without a TTL are often load-bearing), so clean-me only audits memory
// usage and the biggest keys using read-only, non-blocking commands.
type Redis struct{}

func (Redis) Name() string { return "redis" }

func (Redis) Available() bool {
	if !run.Look("redis-cli") {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	out, err := run.Capture(ctx, "redis-cli", "ping")
	return err == nil && strings.TrimSpace(out) == "PONG"
}

var memRe = regexp.MustCompile(`(?m)^used_memory:(\d+)`)

func (Redis) Scan(ctx context.Context, o stage.Options) ([]stage.Item, error) {
	var items []stage.Item

	if info, err := run.Capture(ctx, "redis-cli", "info", "memory"); err == nil {
		var used int64
		if m := memRe.FindStringSubmatch(info); m != nil {
			used = parseDockerSize(m[1]) // plain integer bytes
		}
		items = append(items, stage.Item{
			Stage: "redis", Path: "redis @ localhost:6379",
			Purpose:    "current memory footprint (see --bigkeys / no-TTL audit below)",
			Size:       used,
			Tier:       stage.TierSafe,
			ReportOnly: true,
		})
	}

	// --bigkeys is SCAN-based and non-blocking; surface its summary lines.
	if out, err := run.Capture(ctx, "redis-cli", "--bigkeys"); err == nil {
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Biggest") && strings.Contains(line, "found") {
				items = append(items, stage.Item{
					Stage: "redis", Path: "largest key",
					Purpose:    collapseSpaces(line),
					Tier:       stage.TierSafe,
					ReportOnly: true,
				})
			}
		}
	}
	return items, nil
}

// Apply is a no-op: Redis is never mutated by clean-me.
func (Redis) Apply(ctx context.Context, items []stage.Item, o stage.Options) (stage.Result, error) {
	return stage.Result{Stage: "redis", Skipped: len(items)}, nil
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
