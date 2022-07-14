package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/adminanalytics"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type analyticsCacheRefreshRoutine struct {
	done   chan struct{}
	db     database.DB
	ctx    context.Context
	cancel context.CancelFunc
}

func newAnalyticsCacheRefreshRoutine(db database.DB) goroutine.BackgroundRoutine {
	ctx, cancel := context.WithCancel(context.Background())
	return &analyticsCacheRefreshRoutine{
		done:   make(chan struct{}),
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (a analyticsCacheRefreshRoutine) Start() {
	// Initially run once:
	adminanalytics.RefreshAnalyticsCache(a.ctx, a.db)

	// And then keep running every 24 hours, or until the routine is stopped.
	for {
		select {
		case <-a.done:
			break
		// Run every 24 hours.
		case <-time.After(24 * time.Hour):
			adminanalytics.RefreshAnalyticsCache(a.ctx, a.db)
		}
	}
}

func (a analyticsCacheRefreshRoutine) Stop() {
	// Stop loop.
	close(a.done)
	// Stop current execution.
	a.cancel()
}
