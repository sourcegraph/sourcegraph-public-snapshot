package cleanup

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type UploadService interface {
	NewJanitor(
		interval time.Duration,
		uploadTimeout time.Duration,
		auditLogMaxAge time.Duration,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverBatchSize int,
		commitResolverMaximumCommitLag time.Duration,
	) goroutine.BackgroundRoutine

	NewUploadResetter(interval time.Duration) *dbworker.Resetter
}
