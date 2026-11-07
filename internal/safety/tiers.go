// Package safety centralizes the rules that stand between a user's selection and
// an actual deletion: tier/force gating, interactive confirmation, and
// pre-execution revalidation.
package safety

import "github.com/ankitaniket/clean-me/internal/models"

// Allowed reports whether an item may be acted on given the selected maximum
// tier and --force. Danger-tier items always require --force. Report-only items
// are never actionable.
func Allowed(it models.Item, o models.Options) bool {
	if it.ReportOnly() {
		return false
	}
	if it.Tier > o.MaxTier {
		return false
	}
	if it.Tier >= models.TierDanger && !o.Force {
		return false
	}
	return true
}
