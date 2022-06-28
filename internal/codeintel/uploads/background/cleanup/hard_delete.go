package cleanup

import (
	"context"
	"sort"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const uploadsBatchSize = 100

func (j *janitor) HandleHardDeleter(ctx context.Context) error {
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
		uploads, totalCount, err := j.dbStore.GetUploads(ctx, options)
		if err != nil {
			return errors.Wrap(err, "dbstore.GetUploads")
		}

		if err := j.deleteBatch(ctx, uploadIDs(uploads)); err != nil {
			return err
		}

		count := len(uploads)
		// log.Debug("Deleted data associated with uploads", "upload_count", count)
		j.metrics.numUploadsPurged.Add(float64(count))

		if count >= totalCount {
			break
		}
	}

	return nil
}

// func (j *janitor) HandleError(err error) {
// 	d.metrics.numErrors.Inc()
// 	log15.Error("Failed to hard delete upload records", "error", err)
// }

func (j *janitor) deleteBatch(ctx context.Context, ids []int) (err error) {
	tx, err := j.dbStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := j.lsifStore.Clear(ctx, ids...); err != nil {
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
