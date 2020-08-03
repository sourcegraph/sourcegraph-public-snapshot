package resetter

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func NewIndexResetter(
	s store.Store,
	resetInterval time.Duration,
	metrics workerutil.ResetterMetrics,
) *workerutil.Resetter {
	return workerutil.NewResetter(store.WorkerutilIndexStore(s), workerutil.ResetterOptions{
		Name:     "index resetter",
		Interval: resetInterval,
		Metrics:  metrics,
	})
}
