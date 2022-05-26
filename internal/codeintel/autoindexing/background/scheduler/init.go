package scheduler

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewScheduler(
	autoindexingSvc *autoindexing.Service,
	dbStore DBStore,
	policyMatcher PolicyMatcher,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &scheduler{
		autoindexingSvc: autoindexingSvc,
		dbStore:         dbStore,
		policyMatcher:   policyMatcher,
	})
}
