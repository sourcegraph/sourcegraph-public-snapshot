package scheduler

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type AutoIndexingService interface {
	NewScheduler(
		interval time.Duration,
		repositoryProcessDelay time.Duration,
		repositoryBatchSize int,
		policyBatchSize int,
	) goroutine.BackgroundRoutine

	NewOnDemandScheduler(
		interval time.Duration,
		batchSize int,
	) goroutine.BackgroundRoutine
}
