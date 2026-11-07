package scanner

import (
	"context"
	"runtime"
	"sync"

	"github.com/ankitaniket/clean-me/internal/models"
)

// Scan runs every stage's read-only Scan concurrently and returns the aggregate
// grouped by category, preserving registry order. Unavailable stages and scan
// errors are recorded on the category rather than failing the whole scan.
func Scan(ctx context.Context, stages []Stage, o models.Options) models.ScanResult {
	cats := make([]models.Category, len(stages))
	var wg sync.WaitGroup
	for i, s := range stages {
		wg.Add(1)
		go func(i int, s Stage) {
			defer wg.Done()
			c := models.Category{
				Stage:       s.Name(),
				Title:       s.Title(),
				Description: s.Description(),
				ReportOnly:  s.ReportOnly(),
				Available:   s.Available(),
			}
			if c.Available {
				if items, err := s.Scan(ctx, o); err != nil {
					c.ScanErr = err.Error()
				} else {
					c.Items = items
				}
			}
			cats[i] = c
		}(i, s)
	}
	wg.Wait()
	return models.ScanResult{OS: runtime.GOOS, Categories: cats}
}
