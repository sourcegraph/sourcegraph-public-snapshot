package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func (b backgroundJob) NewCommitGraphUpdater(interval time.Duration, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.uploadSvc.UpdateAllDirtyCommitGraphs(ctx, maxAgeForNonStaleBranches, maxAgeForNonStaleTags)
	}))
}
