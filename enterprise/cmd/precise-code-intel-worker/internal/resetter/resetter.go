package resetter

import (
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func NewUploadResetter(
	s store.Store,
	resetInterval time.Duration,
	metrics workerutil.ResetterMetrics,
) *workerutil.Resetter {
	return workerutil.NewResetter(store.WorkerutilUploadStore(s), workerutil.ResetterOptions{
		Name:     "upload resetter",
		Interval: resetInterval,
		Metrics:  metrics,
	})
}
