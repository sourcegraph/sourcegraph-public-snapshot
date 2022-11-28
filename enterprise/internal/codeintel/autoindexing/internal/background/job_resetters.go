package background

import (
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewIndexResetter returns a background routine that periodically resets index
// records that are marked as being processed but are no longer being processed
// by a worker.
func NewIndexResetter(interval time.Duration, store dbworkerstore.Store[types.Index], logger log.Logger, metrics *resetterMetrics) *dbworker.Resetter[types.Index] {
	return dbworker.NewResetter(logger.Scoped("indexResetter", ""), store, dbworker.ResetterOptions{
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
func NewDependencyIndexResetter(interval time.Duration, store dbworkerstore.Store[shared.DependencyIndexingJob], logger log.Logger, metrics *resetterMetrics) *dbworker.Resetter[shared.DependencyIndexingJob] {
	return dbworker.NewResetter(logger.Scoped("dependencyIndexResetter", ""), store, dbworker.ResetterOptions{
		Name:     "precise_code_intel_dependency_index_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numDependencyIndexResets,
			RecordResetFailures: metrics.numDependencyIndexResetFailures,
			Errors:              metrics.numDependencyIndexResetErrors,
		},
	})
}
