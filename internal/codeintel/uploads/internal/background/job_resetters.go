package background

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

// NewUploadResetter returns a background routine that periodically resets upload
// records that are marked as being processed but are no longer being processed
// by a worker.
func (b backgroundJob) NewUploadResetter(interval time.Duration) *dbworker.Resetter {
	return dbworker.NewResetter(b.logger, b.uploadSvc.GetWorkerutilStore(), dbworker.ResetterOptions{
		Name:     "precise_code_intel_upload_worker_resetter",
		Interval: interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        b.resetterMetrics.numUploadResets,
			RecordResetFailures: b.resetterMetrics.numUploadResetFailures,
			Errors:              b.resetterMetrics.numUploadResetErrors,
		},
	})
}
