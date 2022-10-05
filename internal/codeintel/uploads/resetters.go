package uploads

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

// NewUploadResetter returns a background routine that periodically resets upload
// records that are marked as being processed but are no longer being processed
// by a worker.
func (s *Service) NewUploadResetter(interval time.Duration) *dbworker.Resetter {
	return dbworker.NewResetter(s.logger, s.workerutilStore, dbworker.ResetterOptions{
		Name:     "precise_code_intel_upload_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        s.resetterMetrics.numUploadResets,
			RecordResetFailures: s.resetterMetrics.numUploadResetFailures,
			Errors:              s.resetterMetrics.numUploadResetErrors,
		},
	})
}
