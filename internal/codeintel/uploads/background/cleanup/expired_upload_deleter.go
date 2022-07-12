package cleanup

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (j *janitor) HandleExpiredUploadDeleter(ctx context.Context) error {
	count, err := j.uploadSvc.SoftDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "SoftDeleteExpiredUploads")
	}
	if count > 0 {
		j.logger.Info("Deleted expired codeintel uploads", log.Int("count", count))
		j.metrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

// func (j *janitor) HandleError(err error) {
// 	j.metrics.numErrors.Inc()
// 	log.Error("Failed to delete expired codeintel uploads", "error", err)
// }
