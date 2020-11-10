package background

import (
	"time"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

// NewUploadResetter returns a background routine that periodically resets upload
// records that are marked as being processed but are no longer held by any Postgres
// transaction.
func NewUploadResetter(s DBStore, interval time.Duration, metrics Metrics) *dbworker.Resetter {
	return dbworker.NewResetter(store.WorkerutilUploadStore(s), dbworker.ResetterOptions{
		Name:     "upload resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.UploadResets,
			RecordResetFailures: metrics.UploadResetFailures,
			Errors:              metrics.Errors,
		},
	})
}

// NewIndexResetter returns a background routine that periodically resets index
// records that are marked as being processed but are no longer held by any Postgres
// transaction.
func NewIndexResetter(s DBStore, interval time.Duration, metrics Metrics) *dbworker.Resetter {
	return dbworker.NewResetter(store.WorkerutilIndexStore(s), dbworker.ResetterOptions{
		Name:     "index resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.IndexResets,
			RecordResetFailures: metrics.IndexResetFailures,
			Errors:              metrics.Errors,
		},
	})
}
