package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCommitGraphUpdater(uploadSvc UploadService, interval time.Duration, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return uploadSvc.UpdateAllDirtyCommitGraphs(ctx, maxAgeForNonStaleBranches, maxAgeForNonStaleTags)
	}))
}
