// Package stages implements the individual cleanup domains. Each type satisfies
// stage.Stage and orchestrates a proven external tool or native command behind
// clean-me's confirm-first safety layer.
package stages

import "github.com/ankitaniket/clean-me/internal/stage"

// gated reports whether an item is allowed to be acted on given the user's
// selected maximum tier and --force. Danger-tier items always require --force.
func gated(it stage.Item, o stage.Options) bool {
	if it.Tier > o.MaxTier {
		return false
	}
	if it.Tier >= stage.TierDanger && !o.Force {
		return false
	}
	return true
}
