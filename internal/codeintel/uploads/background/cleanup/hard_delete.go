package cleanup

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (j *janitor) HandleHardDeleter(ctx context.Context) error {
	count, err := j.uploadSvc.HardDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "uploadSvc.HardDeleteExpiredUploads")
	}

	j.metrics.numUploadsPurged.Add(float64(count))

	return nil
}
