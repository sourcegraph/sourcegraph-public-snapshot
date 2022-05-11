package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type auditLogJanitor struct {
	dbStore DBStore
	maxAge  time.Duration
	metrics *metrics
}

var _ goroutine.Handler = &auditLogJanitor{}
var _ goroutine.ErrorHandler = &auditLogJanitor{}

// NewAuditLogJanitor returns a background routine that periodically deletes old audit log records.
func NewAuditLogJanitor(dbStore DBStore, maxAge time.Duration, interval time.Duration, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &auditLogJanitor{
		dbStore: dbStore,
		maxAge:  maxAge,
		metrics: metrics,
	})
}

func (j *auditLogJanitor) Handle(ctx context.Context) (err error) {
	count, err := j.dbStore.DeleteOldAuditLogs(ctx, j.maxAge, time.Now())
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteOldAuditLogs")
	}

	j.metrics.numAuditLogRecordsExpired.Add(float64(count))
	return nil
}

func (j *auditLogJanitor) HandleError(err error) {
	j.metrics.numErrors.Inc()
	log15.Error("Failed to delete codeintel audit log records", "error", err)
}
