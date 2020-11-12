package background

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type AbandonedUploadJanitor struct {
	dbStore DBStore
	ttl     time.Duration
	metrics Metrics
}

var _ goroutine.Handler = &AbandonedUploadJanitor{}

// NewAbandonedUploadJanitor returns a background routine that periodically removes
// upload records which have not left the uploading state within the given TTL.
func NewAbandonedUploadJanitor(dbStore DBStore, ttl, interval time.Duration, metrics Metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &AbandonedUploadJanitor{
		dbStore: dbStore,
		ttl:     ttl,
		metrics: metrics,
	})
}

func (h *AbandonedUploadJanitor) Handle(ctx context.Context) error {
	count, err := h.dbStore.DeleteUploadsStuckUploading(ctx, time.Now().UTC().Add(-h.ttl))
	if err != nil {
		return errors.Wrap(err, "DeleteUploadsStuckUploading")
	}
	if count > 0 {
		log15.Debug("Deleted abandoned upload records", "count", count)
		h.metrics.UploadRecordsRemoved.Add(float64(count))
	}

	return nil
}

func (h *AbandonedUploadJanitor) HandleError(err error) {
	h.metrics.Errors.Inc()
	log15.Error("Failed to delete abandoned uploads", "error", err)
}
