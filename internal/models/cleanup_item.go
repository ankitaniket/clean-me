// Package models holds clean-me's pure data types. It imports nothing from the
// rest of the project so every layer (scanner, cleaner, safety, tui, cli) can
// share these types without creating import cycles.
package models

// Tier expresses how risky an item is to reclaim. The engine only acts on items
// up to the user-selected maximum tier, and the danger tier additionally
// requires --force.
type Tier int

const (
	// TierSafe: re-downloadable / regenerated data (package caches, dangling
	// Docker layers, __pycache__). Included by default; safe to delete.
	TierSafe Tier = 1
	// TierOptIn: unused but not trivially recreated (all unused Docker images,
	// old kernels). Opt-in.
	TierOptIn Tier = 2
	// TierDanger: can destroy real state (Docker volumes, virtualenvs, /tmp).
	// Requires max tier plus --force, and is never preselected in the UI.
	TierDanger Tier = 3
)

func (t Tier) String() string {
	switch t {
	case TierSafe:
		return "safe"
	case TierOptIn:
		return "opt-in"
	case TierDanger:
		return "danger"
	default:
		return "unknown"
	}
}

// Action tells the cleaner how to reclaim an item. Stages only describe the
// action; cleaner/executor.go carries it out. This is what keeps the engine
// independent of any front end.
type Action int

const (
	// ActionTrash moves Item.Path to the trash (recoverable). Used for
	// node_modules, __pycache__, virtualenvs, tool caches.
	ActionTrash Action = iota
	// ActionCommand runs Item.Command (permanent — the underlying tool has no
	// trash). Used for docker prune, brew/apt/dnf clean, npm/pip cache.
	ActionCommand
	// ActionReport is audit-only: the cleaner never touches it. Used for git
	// branch audits and Redis memory reports.
	ActionReport
)

// Item is a single thing clean-me can report on or reclaim. Every destructive
// action is described by an Item so the UI can always show path, purpose and
// reclaimable size before anything happens.
type Item struct {
	Stage       string `json:"stage"`
	Path        string `json:"path"`
	Purpose     string `json:"purpose"`
	Size        int64  `json:"size"`        // logical/apparent size in bytes
	Reclaimable int64  `json:"reclaimable"` // estimated bytes freed (may be < Size for pnpm/bun links)
	Warn        string `json:"warn,omitempty"`
	Tier        Tier   `json:"tier"`
	Action      Action `json:"action"`
	Command     string `json:"command,omitempty"` // for ActionCommand
	Permanent   bool   `json:"permanent"`         // true when deletion is not trash-recoverable
}

// ReportOnly reports whether this item is audit-only.
func (i Item) ReportOnly() bool { return i.Action == ActionReport }

// Key is a stable identifier for an item, used by the UI to track selection
// across navigation.
func (i Item) Key() string { return i.Stage + "\x00" + i.Path }

// Options carries the user's intent into scanning and cleaning.
type Options struct {
	Force   bool
	Yes     bool // non-interactive: proceed without prompting (CI)
	NoTrash bool // permanently delete instead of moving to trash
	MaxTier Tier
	Paths   []string // roots for node/python/git scans
	DryRun  bool
}
