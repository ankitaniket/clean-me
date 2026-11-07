package models

// Category groups one stage's scan output for display in the dashboard and
// category explorer.
type Category struct {
	Stage       string
	Title       string
	Description string
	ReportOnly  bool
	Available   bool
	Items       []Item
	ScanErr     string
}

// Reclaimable is the sum of estimated reclaimable bytes for actionable items.
func (c Category) Reclaimable() int64 {
	var sum int64
	for _, it := range c.Items {
		if !it.ReportOnly() {
			sum += it.Reclaimable
		}
	}
	return sum
}

// ScanResult is the aggregate of every stage's scan.
type ScanResult struct {
	OS         string
	Categories []Category
}

// TotalReclaimable sums reclaimable bytes across all categories.
func (s ScanResult) TotalReclaimable() int64 {
	var sum int64
	for _, c := range s.Categories {
		sum += c.Reclaimable()
	}
	return sum
}

// OpportunityCount counts actionable (non-report) items across all categories.
func (s ScanResult) OpportunityCount() int {
	n := 0
	for _, c := range s.Categories {
		for _, it := range c.Items {
			if !it.ReportOnly() {
				n++
			}
		}
	}
	return n
}

// AllItems flattens actionable items across categories (used by Smart Cleanup).
func (s ScanResult) AllItems() []Item {
	var out []Item
	for _, c := range s.Categories {
		out = append(out, c.Items...)
	}
	return out
}

// ItemOutcome is the result of attempting to reclaim one item.
type ItemOutcome struct {
	Item      Item
	Reclaimed int64
	TrashedTo string
	Err       error
}

// Result aggregates the outcomes of a cleaning run.
type Result struct {
	Outcomes []ItemOutcome
}

// Freed returns total bytes actually reclaimed.
func (r Result) Freed() int64 {
	var sum int64
	for _, o := range r.Outcomes {
		if o.Err == nil {
			sum += o.Reclaimed
		}
	}
	return sum
}

// Trashed returns how many items were moved to the trash (recoverable).
func (r Result) Trashed() int {
	n := 0
	for _, o := range r.Outcomes {
		if o.Err == nil && o.TrashedTo != "" {
			n++
		}
	}
	return n
}

// Failed returns how many item operations errored.
func (r Result) Failed() int {
	n := 0
	for _, o := range r.Outcomes {
		if o.Err != nil {
			n++
		}
	}
	return n
}
