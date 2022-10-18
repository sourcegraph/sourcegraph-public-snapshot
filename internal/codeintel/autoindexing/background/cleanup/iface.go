package cleanup

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type AutoIndexingServiceBackgroundJobs interface {
	NewJanitor(
		interval time.Duration,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverBatchSize int,
		commitResolverMaximumCommitLag time.Duration,
		failedIndexBatchSize int,
		failedIndexMaxAge time.Duration,
	) goroutine.BackgroundRoutine

	NewIndexResetter(interval time.Duration) *dbworker.Resetter
	NewDependencyIndexResetter(interval time.Duration) *dbworker.Resetter
}
