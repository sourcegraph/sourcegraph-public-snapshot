package processor

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewUploadResetter returns a background routine that periodically resets upload
// records that are marked as being processed but are no longer being processed
// by a worker.
func NewUploadResetter(logger log.Logger, store store.Store[shared.Upload], config *Config, metrics *resetterMetrics) *dbworker.Resetter[shared.Upload] {
	return dbworker.NewResetter(logger.Scoped("uploadResetter", ""), store, dbworker.ResetterOptions{
		Name:     "precise_code_intel_upload_worker_resetter",
		Interval: config.Interval,
		Metrics: dbworker.ResetterMetrics{
			RecordResets:        metrics.numUploadResets,
			RecordResetFailures: metrics.numUploadResetFailures,
			Errors:              metrics.numUploadResetErrors,
		},
	})
}
