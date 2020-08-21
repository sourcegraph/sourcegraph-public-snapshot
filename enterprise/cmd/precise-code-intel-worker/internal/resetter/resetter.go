package resetter

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewUploadResetter(
	s store.Store,
	resetInterval time.Duration,
	metrics dbworker.ResetterMetrics,
) goroutine.BackgroundRoutine {
	return dbworker.NewResetter(store.WorkerutilUploadStore(s), dbworker.ResetterOptions{
		Name:     "upload resetter",
		Interval: resetInterval,
		Metrics:  metrics,
	})
}
