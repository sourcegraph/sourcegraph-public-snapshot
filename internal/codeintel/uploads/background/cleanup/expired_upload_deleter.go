package cleanup

import (
	"context"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (j *janitor) HandleExpiredUploadDeleter(ctx context.Context) error {
	count, err := j.dbStore.SoftDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "SoftDeleteExpiredUploads")
	}
	if count > 0 {
		log15.Info("Deleted expired codeintel uploads", "count", count)
		j.metrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

// func (j *janitor) HandleError(err error) {
// 	j.metrics.numErrors.Inc()
// 	log15.Error("Failed to delete expired codeintel uploads", "error", err)
// }
