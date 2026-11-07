// Package stage defines the common contract every cleanup stage implements
// and the data types that flow between scanning, confirmation and reporting.
package stage

import "context"

// Tier expresses how risky an item is to delete. The CLI only acts on items
// up to the user-selected maximum tier, and higher tiers demand --force.
type Tier int

const (
	// TierSafe is re-downloadable / regenerated data: package caches, dangling
	// Docker layers, __pycache__. Reclaiming it is cheap and reversible in
	// practice. Included by default.
	TierSafe Tier = 1
	// TierOptIn is data that is unused but not trivially re-created, e.g. all
	// unused Docker images (not just dangling) or old kernels. Opt-in.
	TierOptIn Tier = 2
	// TierDanger can destroy real state: Docker volumes, virtualenvs, /tmp.
	// Requires the maximum tier plus --force.
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

// Item is a single thing clean-me can report on or reclaim. Every destructive
// action is described by an Item so the confirmation layer can always show the
// user the path, the human-readable purpose and the reclaimable size first.
type Item struct {
	Stage       string `json:"stage"`
	Path        string `json:"path"`
	Purpose     string `json:"purpose"`     // why this is safe/removable, in plain words
	Size        int64  `json:"size"`        // apparent size in bytes
	Reclaimable int64  `json:"reclaimable"` // bytes actually freed (may be < Size for pnpm/bun links)
	Warn        string `json:"warn,omitempty"`
	Tier        Tier   `json:"tier"`
	ReportOnly  bool   `json:"reportOnly"` // true for git/redis: audited, never deleted
	Command     string `json:"command,omitempty"`
}

// Options carries the user's intent from the command line into Apply.
type Options struct {
	DryRun  bool
	Force   bool
	Yes     bool // non-interactive: proceed without prompting (CI)
	NoTrash bool // permanently delete instead of moving to trash
	MaxTier Tier
	Paths   []string // roots for node/python/git scans
}

// Result summarizes what a stage actually did during Apply.
type Result struct {
	Stage     string
	Reclaimed int64 // bytes actually freed
	Removed   int   // items acted on
	Skipped   int   // items left alone (declined / gated by tier/force)
	Errors    []error
}

// Stage is one cleanup domain (docker, system, node, ...). Scan is always
// read-only and builds the preview; Apply is the only place deletion happens.
type Stage interface {
	Name() string
	Available() bool
	Scan(ctx context.Context, o Options) ([]Item, error)
	Apply(ctx context.Context, items []Item, o Options) (Result, error)
}
