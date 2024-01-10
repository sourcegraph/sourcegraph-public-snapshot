package dependencies

import (
	"time"

	"github.com/sourcegraph/log"

	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewIndexResetter returns a background routine that periodically resets index
// records that are marked as being processed but are no longer being processed
// by a worker.
func NewIndexResetter(logger log.Logger, interval time.Duration, store dbworkerstore.Store[uploadsshared.Index], metrics *resetterMetrics) *dbworker.Resetter[uploadsshared.Index] {
	return dbworker.NewResetter(logger.Scoped("indexResetter"), store, dbworker.ResetterOptions{
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
func NewDependencyIndexResetter(logger log.Logger, interval time.Duration, store dbworkerstore.Store[dependencyIndexingJob], metrics *resetterMetrics) *dbworker.Resetter[dependencyIndexingJob] {
	return dbworker.NewResetter(logger.Scoped("dependencyIndexResetter"), store, dbworker.ResetterOptions{
		Name:     "precise_code_intel_dependency_index_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numDependencyIndexResets,
			RecordResetFailures: metrics.numDependencyIndexResetFailures,
			Errors:              metrics.numDependencyIndexResetErrors,
		},
	})
}
