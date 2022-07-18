package cleanup

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (j *janitor) HandleAuditLog(ctx context.Context) (err error) {
	count, err := j.uploadSvc.DeleteOldAuditLogs(ctx, ConfigInst.AuditLogMaxAge, time.Now())
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteOldAuditLogs")
	}

	j.metrics.numAuditLogRecordsExpired.Add(float64(count))
	return nil
}

// func (j *janitor) HandleError(err error) {
// 	j.metrics.numErrors.Inc()
// 	log15.Error("Failed to delete codeintel audit log records", "error", err)
// }
