package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCVEMatcher(store store.Store, metrics *Metrics, interval time.Duration) goroutine.BackgroundRoutine {
	cveMatcher := &cveMatcher{
		store: store,
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.sentinel-cve-matcher", "TODO",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return cveMatcher.handle(ctx, metrics)
		}),
	)
}

type cveMatcher struct {
	store store.Store
}

func (matcher *cveMatcher) handle(ctx context.Context, metrics *Metrics) error {
	if err := matcher.store.ScanMatches(ctx); err != nil {
		return err
	}

	return nil
}
