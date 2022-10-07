package backfill

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type UploadService interface {
	NewCommittedAtBackfiller(
		interval time.Duration,
		batchSize int,
	) goroutine.BackgroundRoutine
}
