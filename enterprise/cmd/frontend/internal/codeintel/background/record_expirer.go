package background

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type RecordExpirer struct {
	dbStore DBStore
	ttl     time.Duration
	metrics Metrics
}

var _ goroutine.Handler = &RecordExpirer{}

// NewRecordExpirer returns a background routine that periodically removes upload
// and index records that are older than the given TTL. Upload records which have
// valid LSIF data (not just a historic upload failure record) will only be deleted
// if it is not visible at the tip of its repository's default branch.
func NewRecordExpirer(dbStore DBStore, ttl, interval time.Duration, metrics Metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &RecordExpirer{
		dbStore: dbStore,
		ttl:     ttl,
		metrics: metrics,
	})
}

func (e *RecordExpirer) Handle(ctx context.Context) error {
	tx, err := e.dbStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	count, err := tx.SoftDeleteOldDumps(ctx, e.ttl, time.Now())
	if err != nil {
		return errors.Wrap(err, "SoftDeleteOldDumps")
	}
	if count > 0 {
		log15.Debug("Deleted old dump records", "count", count)
		e.metrics.UploadRecordsRemoved.Add(float64(count))
	}

	// TODO (efritz)- expire old upload (non-dump) records
	// TODO (efritz)- expire old index records
	return nil
}

func (e *RecordExpirer) HandleError(err error) {
	e.metrics.Errors.Inc()
	log15.Error("Failed to delete old codeintel records", "error", err)
}
