package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type abandonedUploadJanitor struct {
	dbStore DBStore
	ttl     time.Duration
	metrics *metrics
}

var _ goroutine.Handler = &abandonedUploadJanitor{}
var _ goroutine.ErrorHandler = &abandonedUploadJanitor{}

// NewAbandonedUploadJanitor returns a background routine that periodically removes
// upload records which have not left the uploading state within the given TTL.
func NewAbandonedUploadJanitor(dbStore DBStore, ttl, interval time.Duration, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &abandonedUploadJanitor{
		dbStore: dbStore,
		ttl:     ttl,
		metrics: metrics,
	})
}

func (h *abandonedUploadJanitor) Handle(ctx context.Context) error {
	count, err := h.dbStore.DeleteUploadsStuckUploading(ctx, time.Now().UTC().Add(-h.ttl))
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteUploadsStuckUploading")
	}
	if count > 0 {
		log15.Debug("Deleted abandoned upload records", "count", count)
		h.metrics.numUploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (h *abandonedUploadJanitor) HandleError(err error) {
	h.metrics.numErrors.Inc()
	log15.Error("Failed to delete abandoned uploads", "error", err)
}
