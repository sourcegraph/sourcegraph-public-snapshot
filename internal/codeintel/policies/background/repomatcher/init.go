package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewMatcher(dbStore DBStore, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &matcher{
		dbStore: dbStore,
		metrics: metrics,
	})
}
