// Package config loads a user config file and named profiles that preset the
// scan scope, so users don't have to retype flags on every run.
//
// Profiles change SCOPE ONLY: which stages run, the max tier, the scan paths,
// the min file size, and extra excludes. Profiles NEVER weaken core safety:
// Apply never touches Force, Yes, or NoTrash. If a user wants those changed,
// they must set the corresponding flags themselves.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ankitaniket/clean-me/internal/models"
)

// Profile presets scan scope. Every field is optional; a zero value means
// "leave the current Options setting untouched".
type Profile struct {
	Stages  []string `json:"stages,omitempty"`  // limit to these stage names
	Tier    int      `json:"tier,omitempty"`    // max tier (1..3); 0 = leave default
	Paths   []string `json:"paths,omitempty"`   // scan roots
	MinSize string   `json:"minSize,omitempty"` // e.g. "200MB"
	Exclude []string `json:"exclude,omitempty"` // extra protected paths
}

// Config is the on-disk user config: a default profile name plus a map of
// user-defined profiles.
type Config struct {
	DefaultProfile string             `json:"defaultProfile,omitempty"`
	Profiles       map[string]Profile `json:"profiles,omitempty"`
}

// Path returns the config file location: <base>/.config/clean-me/config.json.
// If base is "", the user's home directory is used.
func Path(base string) string {
	if base == "" {
		if home, err := os.UserHomeDir(); err == nil {
			base = home
		}
	}
	return filepath.Join(base, ".config", "clean-me", "config.json")
}

// Load reads and parses the config file. A missing file yields an empty Config
// and a nil error (it is not an error to have no config).
func Load(base string) (*Config, error) {
	c := &Config{}
	data, err := os.ReadFile(Path(base))
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return c, nil
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Save writes the config as pretty JSON with 0600 perms, creating the parent
// directory if needed.
func Save(base string, c *Config) error {
	p := Path(base)
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

// home returns the user's home directory, or "~" if it cannot be determined.
func home() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "~"
}

// Builtin returns the preset profiles shipped with clean-me. These are used as
// a fallback when a profile name is not present in the user's config. "~" path
// entries are expanded to the real home directory.
func Builtin() map[string]Profile {
	h := home()
	return map[string]Profile{
		// Laptop: broad opt-in scope across the whole home directory.
		"laptop": {
			Tier:  int(models.TierOptIn),
			Paths: []string{h},
		},
		// Server: conservative, safe-tier only, no largefiles/xcode/trash.
		"server": {
			Stages: []string{"system", "docker", "node", "python"},
			Tier:   int(models.TierSafe),
			Paths:  []string{"."},
		},
		// Docker-focused cleanup.
		"docker": {
			Stages: []string{"docker", "system"},
			Tier:   int(models.TierOptIn),
		},
		// Node-focused cleanup, safe tier, scoped to home.
		"node": {
			Stages: []string{"node", "caches"},
			Tier:   int(models.TierSafe),
			Paths:  []string{h},
		},
		// Python-focused cleanup, safe tier, scoped to home.
		"python": {
			Stages: []string{"python", "caches"},
			Tier:   int(models.TierSafe),
			Paths:  []string{h},
		},
	}
}

// Resolve looks up a profile by name: the user's config takes precedence, and
// the builtin presets are the fallback. The bool reports whether a profile was
// found.
func Resolve(c *Config, name string) (Profile, bool) {
	if c != nil {
		if p, ok := c.Profiles[name]; ok {
			return p, true
		}
	}
	if p, ok := Builtin()[name]; ok {
		return p, true
	}
	return Profile{}, false
}

// Apply mutates Options from a Profile and returns the stage filter (p.Stages)
// so the caller can apply it.
//
// It sets MaxTier only when p.Tier > 0, sets Paths only when non-empty, sets
// MinFileSize via parseSize when p.MinSize is non-empty, and appends p.Exclude
// to o.Exclude. It NEVER touches Force, Yes, or NoTrash: profiles change scope,
// not core safety.
func Apply(p Profile, o *models.Options, parseSize func(string) int64) []string {
	if o == nil {
		return p.Stages
	}
	if p.Tier > 0 {
		o.MaxTier = models.Tier(p.Tier)
	}
	if len(p.Paths) > 0 {
		o.Paths = p.Paths
	}
	if p.MinSize != "" && parseSize != nil {
		o.MinFileSize = parseSize(p.MinSize)
	}
	if len(p.Exclude) > 0 {
		o.Exclude = append(o.Exclude, p.Exclude...)
	}
	return p.Stages
}
