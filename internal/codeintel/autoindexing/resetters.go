package autoindexing

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

// NewIndexResetter returns a background routine that periodically resets index
// records that are marked as being processed but are no longer being processed
// by a worker.
func (s *Service) NewIndexResetter(interval time.Duration) *dbworker.Resetter {
	return dbworker.NewResetter(s.logger, s.workerutilStore, dbworker.ResetterOptions{
		Name:     "precise_code_intel_index_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        s.metrics.numIndexResets,
			RecordResetFailures: s.metrics.numIndexResetFailures,
			Errors:              s.metrics.numIndexResetErrors,
		},
	})
}

// NewDependencyIndexResetter returns a background routine that periodically resets
// dependency index records that are marked as being processed but are no longer being
// processed by a worker.
func (s *Service) NewDependencyIndexResetter(interval time.Duration) *dbworker.Resetter {
	return dbworker.NewResetter(s.logger, s.dependencyIndexingStore, dbworker.ResetterOptions{
		Name:     "precise_code_intel_dependency_index_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        s.metrics.numDependencyIndexResets,
			RecordResetFailures: s.metrics.numDependencyIndexResetFailures,
			Errors:              s.metrics.numDependencyIndexResetErrors,
		},
	})
}
