package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func (b backgroundJob) NewUploadExpirer(
	interval time.Duration,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	uploadProcessDelay time.Duration,
	uploadBatchSize int,
	commitBatchSize int,
	policyBatchSize int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.uploadSvc.HandleExpiredUploadsBatch(ctx, b.expirationMetrics, ExpirerConfig{
			RepositoryProcessDelay: repositoryProcessDelay,
			RepositoryBatchSize:    repositoryBatchSize,
			UploadProcessDelay:     uploadProcessDelay,
			UploadBatchSize:        uploadBatchSize,
			CommitBatchSize:        commitBatchSize,
			PolicyBatchSize:        policyBatchSize,
		})
	}))
}

type ExpirerConfig struct {
	RepositoryProcessDelay time.Duration
	RepositoryBatchSize    int
	UploadProcessDelay     time.Duration
	UploadBatchSize        int
	CommitBatchSize        int
	PolicyBatchSize        int
}
