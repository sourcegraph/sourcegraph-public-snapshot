package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCVEDownloader(store store.Store, metrics *Metrics, interval time.Duration) goroutine.BackgroundRoutine {
	cveDownloader := &CveDownloader{
		store: store,
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.sentinel-cve-downloader", "TODO",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			vulnerabilities, err := cveDownloader.handle(ctx, metrics)
			if err != nil {
				return err
			}

			if err := store.InsertVulnerabilities(ctx, vulnerabilities); err != nil {
				return err
			}

			return nil
		}),
	)
}

type CveDownloader struct {
	store store.Store
}

func (matcher *CveDownloader) handle(ctx context.Context, metrics *Metrics) (vulns []shared.Vulnerability, err error) {
	return ReadGitHubAdvisoryDB(ctx, false)
}
