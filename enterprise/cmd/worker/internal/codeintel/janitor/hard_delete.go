package janitor

import (
	"context"
	"sort"
	"time"

	"github.com/inconshreveable/log15"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type hardDeleter struct {
	dbStore   DBStore
	lsifStore LSIFStore
	metrics   *metrics
}

var (
	_ goroutine.Handler      = &hardDeleter{}
	_ goroutine.ErrorHandler = &hardDeleter{}
)

// NewHardDeleter returns a background routine that periodically hard-deletes all
// soft-deleted upload records. Each upload record marked as soft-deleted in the
// database will have its associated data in the code intel deleted, and the upload
// record hard-deleted.
//
// This cleanup routine subsumes an old routine that would remove any records which
// did not have an associated upload record. Doing a soft-delete and a transactional
// cleanup routine instead ensures we delete unreachable data as soon as it's no longer
// referenceable.
func NewHardDeleter(dbStore DBStore, lsifStore LSIFStore, interval time.Duration, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &hardDeleter{
		dbStore:   dbStore,
		lsifStore: lsifStore,
		metrics:   metrics,
	})
}

const uploadsBatchSize = 100

func (d *hardDeleter) Handle(ctx context.Context) error {
	options := store.GetUploadsOptions{
		State:            "deleted",
		Limit:            uploadsBatchSize,
		AllowExpired:     true,
		AllowDeletedRepo: true,
	}

	for {
		// Always request the first page of deleted uploads. If this is not
		// the first iteration of the loop, then the previous iteration has
		// deleted the records that composed the previous page, and the
		// previous "second" page is now the first page.
		uploads, totalCount, err := d.dbStore.GetUploads(ctx, options)
		if err != nil {
			return errors.Wrap(err, "dbstore.GetUploads")
		}

		if err := d.deleteBatch(ctx, uploadIDs(uploads)); err != nil {
			return err
		}

		count := len(uploads)
		log15.Debug("Deleted data associated with uploads", "upload_count", count)
		d.metrics.numUploadsPurged.Add(float64(count))

		if count >= totalCount {
			break
		}
	}

	return nil
}

func (d *hardDeleter) HandleError(err error) {
	d.metrics.numErrors.Inc()
	log15.Error("Failed to hard delete upload records", "error", err)
}

func (d *hardDeleter) deleteBatch(ctx context.Context, ids []int) (err error) {
	tx, err := d.dbStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := d.lsifStore.Clear(ctx, ids...); err != nil {
		return errors.Wrap(err, "lsifstore.Clear")
	}

	if err := tx.HardDeleteUploadByID(ctx, ids...); err != nil {
		return errors.Wrap(err, "dbstore.HardDeleteUploadByID")
	}

	return nil
}

func uploadIDs(uploads []store.Upload) []int {
	ids := make([]int, 0, len(uploads))
	for i := range uploads {
		ids = append(ids, uploads[i].ID)
	}
	sort.Ints(ids)

	return ids
}
