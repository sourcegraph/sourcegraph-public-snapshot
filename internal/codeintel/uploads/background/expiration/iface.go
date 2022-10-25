package expiration

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type UploadServiceBackgroundJobs interface {
	NewUploadExpirer(
		interval time.Duration,
		repositoryProcessDelay time.Duration,
		repositoryBatchSize int,
		uploadProcessDelay time.Duration,
		uploadBatchSize int,
		commitBatchSize int,
		policyBatchSize int,
	) goroutine.BackgroundRoutine
}
