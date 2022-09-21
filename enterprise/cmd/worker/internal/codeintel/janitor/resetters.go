package janitor

import (
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewUploadResetter returns a background routine that periodically resets upload
// records that are marked as being processed but are no longer being processed
// by a worker.
func NewUploadResetter(logger log.Logger, s dbworkerstore.Store, interval time.Duration, metrics *metrics) *dbworker.Resetter {
	return dbworker.NewResetter(logger.Scoped("dbworker.Resetter", ""), s, dbworker.ResetterOptions{
		Name:     "precise_code_intel_upload_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numUploadResets,
			RecordResetFailures: metrics.numUploadResetFailures,
			Errors:              metrics.numUploadResetErrors,
		},
	})
}

// NewIndexResetter returns a background routine that periodically resets index
// records that are marked as being processed but are no longer being processed
// by a worker.
func NewIndexResetter(logger log.Logger, s dbworkerstore.Store, interval time.Duration, metrics *metrics) *dbworker.Resetter {
	return dbworker.NewResetter(logger.Scoped("dbworker.Resetter", ""), s, dbworker.ResetterOptions{
		Name:     "precise_code_intel_index_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numIndexResets,
			RecordResetFailures: metrics.numIndexResetFailures,
			Errors:              metrics.numIndexResetErrors,
		},
	})
}

// NewDependencyIndexResetter returns a background routine that periodically resets
// dependency index records that are marked as being processed but are no longer being
// processed by a worker.
func NewDependencyIndexResetter(logger log.Logger, s dbworkerstore.Store, interval time.Duration, metrics *metrics) *dbworker.Resetter {
	return dbworker.NewResetter(logger.Scoped("dbworker.Resetter", ""), s, dbworker.ResetterOptions{
		Name:     "precise_code_intel_dependency_index_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numDependencyIndexResets,
			RecordResetFailures: metrics.numDependencyIndexResetFailures,
			Errors:              metrics.numDependencyIndexResetErrors,
		},
	})
}
