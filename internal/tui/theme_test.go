package tui

import (
	"strings"
	"testing"
)

// projectLabel must stay unambiguous: a node_modules under a generic subfolder
// still surfaces the project name, while other paths keep their last 2 segments.
func TestProjectLabel(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		contains string
	}{
		{"direct node_modules", "/Users/you/Projects/scene-forge/node_modules", "scene-forge"},
		{"generic subfolder node_modules", "/Users/you/Projects/realtime-video-agent/web/node_modules", "realtime-video-agent"},
		{"plain cache path", "/Users/you/.npm/_cacache", ".npm/_cacache"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := projectLabel(c.path)
			if !strings.Contains(got, c.contains) {
				t.Errorf("projectLabel(%q) = %q, want it to contain %q", c.path, got, c.contains)
			}
		})
	}
}
