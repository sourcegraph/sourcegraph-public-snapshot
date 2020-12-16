package background

import (
	"time"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

// NewUploadResetter returns a background routine that periodically resets upload
// records that are marked as being processed but are no longer held by any Postgres
// transaction.
func NewUploadResetter(s DBStore, interval time.Duration, operations *operations) *dbworker.Resetter {
	return dbworker.NewResetter(store.WorkerutilUploadStore(s), dbworker.ResetterOptions{
		Name:     "precise_code_intel_upload_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        operations.numUploadResets,
			RecordResetFailures: operations.numUploadResetFailures,
			Errors:              operations.numErrors,
		},
	})
}

// NewIndexResetter returns a background routine that periodically resets index
// records that are marked as being processed but are no longer held by any Postgres
// transaction.
func NewIndexResetter(s DBStore, interval time.Duration, operations *operations) *dbworker.Resetter {
	return dbworker.NewResetter(store.WorkerutilIndexStore(s), dbworker.ResetterOptions{
		Name:     "precise_code_intel_index_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        operations.numIndexResets,
			RecordResetFailures: operations.numIndexResetFailures,
			Errors:              operations.numErrors,
		},
	})
}
