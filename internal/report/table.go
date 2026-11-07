// Package report renders the scan preview: the grouped "path · purpose · size"
// table and reclaimable-space totals that the user sees before confirming.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/ankitaniket/clean-me/internal/stage"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
	totalStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

// Humanize renders a byte count as a short human-readable string.
func Humanize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// Text writes the grouped preview table grouped by stage. Returns total
// reclaimable bytes across all non-report items.
func Text(w io.Writer, items []stage.Item) int64 {
	byStage := map[string][]stage.Item{}
	order := []string{}
	for _, it := range items {
		if _, ok := byStage[it.Stage]; !ok {
			order = append(order, it.Stage)
		}
		byStage[it.Stage] = append(byStage[it.Stage], it)
	}
	sort.Strings(order)

	var grand int64
	for _, name := range order {
		group := byStage[name]
		var sub int64
		for _, it := range group {
			if !it.ReportOnly {
				sub += it.Reclaimable
			}
		}
		grand += sub

		header := titleStyle.Render(fmt.Sprintf("▸ %s", strings.ToUpper(name)))
		if group[0].ReportOnly {
			header += dimStyle.Render("  (report-only — nothing is deleted)")
		} else {
			header += dimStyle.Render(fmt.Sprintf("  (%s reclaimable)", Humanize(sub)))
		}
		fmt.Fprintln(w, "\n"+header)

		tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
		fmt.Fprintln(tw, "  TIER\tSIZE\tPURPOSE\tPATH")
		for _, it := range group {
			size := Humanize(it.Reclaimable)
			if it.ReportOnly {
				size = Humanize(it.Size)
			}
			fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n", it.Tier, size, it.Purpose, it.Path)
			if it.Warn != "" {
				fmt.Fprintf(tw, "  \t\t%s\t\n", warnStyle.Render("⚠ "+it.Warn))
			}
		}
		tw.Flush()
	}

	fmt.Fprintln(w, "\n"+totalStyle.Render(fmt.Sprintf("Total reclaimable: %s", Humanize(grand))))
	return grand
}

// JSON writes the items as machine-readable JSON.
func JSON(w io.Writer, items []stage.Item) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}
