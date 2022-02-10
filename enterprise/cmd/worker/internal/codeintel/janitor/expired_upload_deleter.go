package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type expiredUploadDeleter struct {
	dbStore DBStore
	metrics *metrics
}

var _ goroutine.Handler = &expiredUploadDeleter{}
var _ goroutine.ErrorHandler = &expiredUploadDeleter{}

// NewExpiredUploadDeleter returns a background routine that periodically
// marks upload records that are both expired and have no references as
// deleted.
//
// The associated repositories will be marked as dirty so that their commit
// graphs are updated in the near future. The data associated with a deleted
// upload is removed in this package's hardDeleter.
func NewExpiredUploadDeleter(dbStore DBStore, interval time.Duration, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &expiredUploadDeleter{
		dbStore: dbStore,
		metrics: metrics,
	})
}

func (e *expiredUploadDeleter) Handle(ctx context.Context) error {
	count, err := e.dbStore.SoftDeleteExpiredUploads(ctx)
	if err != nil {
		return errors.Wrap(err, "SoftDeleteExpiredUploads")
	}
	if count > 0 {
		log15.Info("Deleted expired codeintel uploads", "count", count)
		e.metrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (e *expiredUploadDeleter) HandleError(err error) {
	e.metrics.numErrors.Inc()
	log15.Error("Failed to delete expired codeintel uploads", "error", err)
}
