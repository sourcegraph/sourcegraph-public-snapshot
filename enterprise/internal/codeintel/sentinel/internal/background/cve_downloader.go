package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCVEDownloader(store store.Store, metrics *Metrics, interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.sentinel-cve-downloader", "TODO",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			// Currently unimplemented
			return nil
		}),
	)
}
