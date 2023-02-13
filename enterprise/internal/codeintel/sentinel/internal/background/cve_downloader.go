package background

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCVEDownloader(store store.Store, metrics *Metrics, interval time.Duration) goroutine.BackgroundRoutine {
	cveDownloader := &cveDownloader{
		store: store,
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.sentinel-cve-downloader", "TODO",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return cveDownloader.handle(ctx, metrics)
		}),
	)
}

type cveDownloader struct {
	store store.Store
}

func (matcher *cveDownloader) handle(ctx context.Context, metrics *Metrics) error {
	// TODO
	fmt.Printf("DOWNLOADER HIT\n")
	return nil
}
