package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ExpirerConfig struct {
	RepositoryProcessDelay time.Duration
	RepositoryBatchSize    int
	UploadProcessDelay     time.Duration
	UploadBatchSize        int
	CommitBatchSize        int
	PolicyBatchSize        int
}

func NewUploadExpirer(
	uploadSvc UploadService,
	interval time.Duration,
	config ExpirerConfig,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return uploadSvc.HandleExpiredUploadsBatch(ctx, NewExpirationMetrics(observationContext), config)
	}))
}
