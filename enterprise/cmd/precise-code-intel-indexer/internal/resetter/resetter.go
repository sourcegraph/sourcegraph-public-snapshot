package resetter

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewIndexResetter(
	s store.Store,
	resetInterval time.Duration,
	metrics dbworker.ResetterMetrics,
) goroutine.BackgroundRoutine {
	return dbworker.NewResetter(store.WorkerutilIndexStore(s), dbworker.ResetterOptions{
		Name:     "index resetter",
		Interval: resetInterval,
		Metrics:  metrics,
	})
}
