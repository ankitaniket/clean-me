package stages

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ankitaniket/clean-me/internal/run"
	"github.com/ankitaniket/clean-me/internal/stage"
)

// Docker reclaims space from the Docker daemon. Because `docker ... prune` has
// no dry-run, the preview sizes come from `docker system df --format json`, and
// the actual reclaim runs the targeted prune after confirmation.
type Docker struct{}

func (Docker) Name() string { return "docker" }

func (Docker) Available() bool {
	if !run.Look("docker") {
		return false
	}
	// Daemon reachable?
	_, err := run.Capture(context.Background(), "docker", "system", "df", "--format", "{{json .}}")
	return err == nil
}

// dfLine mirrors one NDJSON record from `docker system df --format json`.
type dfLine struct {
	Type        string `json:"Type"`
	TotalCount  string `json:"TotalCount"`
	Active      string `json:"Active"`
	Size        string `json:"Size"`
	Reclaimable string `json:"Reclaimable"` // e.g. "1.093GB (40%)"
}

func (Docker) Scan(ctx context.Context, o stage.Options) ([]stage.Item, error) {
	out, err := run.Capture(ctx, "docker", "system", "df", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}
	var items []stage.Item
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var d dfLine
		if err := json.Unmarshal([]byte(line), &d); err != nil {
			continue
		}
		recl := parseDockerSize(d.Reclaimable)
		if recl <= 0 {
			continue
		}
		it := stage.Item{Stage: "docker", Size: recl, Reclaimable: recl}
		switch d.Type {
		case "Images":
			it.Path = "docker images (unused)"
			it.Purpose = "unused images, incl. dangling layers"
			it.Tier = stage.TierOptIn
			it.Command = "docker image prune -a -f --filter until=168h"
			it.Warn = "tier 2: removes all images not used by a container (may require re-pull)"
		case "Containers":
			it.Path = "docker containers (stopped)"
			it.Purpose = "stopped containers holding writable layers"
			it.Tier = stage.TierSafe
			it.Command = "docker container prune -f"
		case "Local Volumes", "Volumes":
			it.Path = "docker volumes (unused)"
			it.Purpose = "volumes not attached to any container — MAY CONTAIN REAL DATA"
			it.Tier = stage.TierDanger
			it.Command = "docker volume prune -f"
			it.Warn = "tier 3: volumes can hold database data; verify before removing"
		case "Build Cache":
			it.Path = "docker build cache"
			it.Purpose = "BuildKit layer cache, safely regenerated on next build"
			it.Tier = stage.TierSafe
			it.Command = "docker builder prune -f --filter until=168h"
		default:
			continue
		}
		items = append(items, it)
	}
	return items, sc.Err()
}

func (Docker) Apply(ctx context.Context, items []stage.Item, o stage.Options) (stage.Result, error) {
	res := stage.Result{Stage: "docker"}
	for _, it := range items {
		if !gated(it, o) {
			res.Skipped++
			continue
		}
		args := strings.Fields(it.Command)
		if _, err := run.Run(ctx, o.DryRun, args[0], args[1:]...); err != nil {
			res.Errors = append(res.Errors, fmt.Errorf("%s: %w", it.Path, err))
			continue
		}
		res.Removed++
		res.Reclaimed += it.Reclaimable // approximate: df-reported reclaimable
	}
	return res, nil
}

// parseDockerSize turns "1.093GB (40%)" / "512MB" / "0B" into bytes.
func parseDockerSize(s string) int64 {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, ' '); i >= 0 {
		s = s[:i] // drop the " (40%)" suffix
	}
	if s == "" {
		return 0
	}
	// split trailing unit letters from the number
	num := s
	unit := ""
	for i := 0; i < len(s); i++ {
		if (s[i] < '0' || s[i] > '9') && s[i] != '.' {
			num, unit = s[:i], s[i:]
			break
		}
	}
	v, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0
	}
	mult := map[string]float64{
		"B": 1, "": 1,
		"kB": 1e3, "KB": 1e3, "KiB": 1024,
		"MB": 1e6, "MiB": 1 << 20,
		"GB": 1e9, "GiB": 1 << 30,
		"TB": 1e12, "TiB": 1 << 40,
	}
	m, ok := mult[strings.TrimSpace(unit)]
	if !ok {
		m = 1
	}
	return int64(v * m)
}
