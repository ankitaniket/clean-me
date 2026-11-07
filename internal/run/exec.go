// Package run wraps external command execution so stages can shell out to the
// proven ecosystem tools (docker, git, apt, brew, du, ...) uniformly, with
// dry-run awareness and captured output.
package run

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Look reports whether a binary is present on PATH.
func Look(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

// Capture runs a command and returns its stdout. Used for read-only queries
// (sizing, listing) during Scan.
func Capture(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return out.String(), fmt.Errorf("%s %s: %s", name, strings.Join(args, " "), msg)
	}
	return out.String(), nil
}

// Run executes a destructive command, honoring dry-run. When dryRun is true it
// does not execute; it returns the command line it would have run so callers
// can surface it to the user.
func Run(ctx context.Context, dryRun bool, name string, args ...string) (string, error) {
	line := name + " " + strings.Join(args, " ")
	if dryRun {
		return "[dry-run] would run: " + strings.TrimSpace(line), nil
	}
	out, err := Capture(ctx, name, args...)
	return out, err
}
