// Package report renders the CLI scan preview: the grouped
// "path · purpose · size" table, humanized sizes, and JSON output.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/ankitaniket/clean-me/internal/models"
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

// Text writes the grouped preview grouped by category and returns total
// reclaimable bytes.
func Text(w io.Writer, res models.ScanResult) int64 {
	for _, c := range res.Categories {
		if !c.Available || len(c.Items) == 0 {
			continue
		}
		header := titleStyle.Render("▸ " + strings.ToUpper(c.Title))
		if c.ReportOnly {
			header += dimStyle.Render("  (report-only — nothing is deleted)")
		} else {
			header += dimStyle.Render(fmt.Sprintf("  (%s reclaimable)", Humanize(c.Reclaimable())))
		}
		fmt.Fprintln(w, "\n"+header)

		tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
		fmt.Fprintln(tw, "  TIER\tSIZE\tPURPOSE\tPATH")
		for _, it := range c.Items {
			size := Humanize(it.Reclaimable)
			if it.ReportOnly() {
				size = Humanize(it.Size)
			}
			fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n", it.Tier, size, it.Purpose, it.Path)
			if it.Warn != "" {
				fmt.Fprintf(tw, "  \t\t%s\t\n", warnStyle.Render("⚠ "+it.Warn))
			}
		}
		tw.Flush()
	}

	total := res.TotalReclaimable()
	fmt.Fprintln(w, "\n"+totalStyle.Render(fmt.Sprintf("Total reclaimable: %s", Humanize(total))))
	return total
}

// JSON writes all scanned items as machine-readable JSON.
func JSON(w io.Writer, res models.ScanResult) error {
	var items []models.Item
	for _, c := range res.Categories {
		items = append(items, c.Items...)
	}
	if items == nil {
		items = []models.Item{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}
