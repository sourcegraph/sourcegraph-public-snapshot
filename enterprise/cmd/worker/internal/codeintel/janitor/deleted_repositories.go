package janitor

import (
	"context"
	"sort"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type deletedRepositoryJanitor struct {
	dbStore DBStore
	metrics *metrics
}

var _ goroutine.Handler = &deletedRepositoryJanitor{}
var _ goroutine.ErrorHandler = &deletedRepositoryJanitor{}

// NewDeletedRepositoryJanitor returns a background routine that periodically
// deletes upload and index records for repositories that have been soft-deleted.
func NewDeletedRepositoryJanitor(dbStore DBStore, interval time.Duration, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &deletedRepositoryJanitor{
		dbStore: dbStore,
		metrics: metrics,
	})
}

func (j *deletedRepositoryJanitor) Handle(ctx context.Context) (err error) {
	tx, err := j.dbStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	uploadsCounts, err := tx.DeleteUploadsWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteUploadsWithoutRepository")
	}

	indexesCounts, err := tx.DeleteIndexesWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "dbstore.DeleteIndexesWithoutRepository")
	}

	for _, counts := range gatherCounts(uploadsCounts, indexesCounts) {
		log15.Debug(
			"Deleted codeintel records with a deleted repository",
			"repository_id", counts.repoID,
			"uploads_count", counts.uploadsCount,
			"indexes_count", counts.indexesCount,
		)

		j.metrics.numUploadRecordsRemoved.Add(float64(counts.uploadsCount))
		j.metrics.numIndexRecordsRemoved.Add(float64(counts.indexesCount))
	}

	return nil
}

func (j *deletedRepositoryJanitor) HandleError(err error) {
	j.metrics.numErrors.Inc()
	log15.Error("Failed to delete codeintel records with a deleted repository", "error", err)
}

type recordCount struct {
	repoID       int
	uploadsCount int
	indexesCount int
}

func gatherCounts(uploadsCounts, indexesCounts map[int]int) []recordCount {
	repoIDsMap := map[int]struct{}{}
	for repoID := range uploadsCounts {
		repoIDsMap[repoID] = struct{}{}
	}
	for repoID := range indexesCounts {
		repoIDsMap[repoID] = struct{}{}
	}

	var repoIDs []int
	for repoID := range repoIDsMap {
		repoIDs = append(repoIDs, repoID)
	}
	sort.Ints(repoIDs)

	recordCounts := make([]recordCount, 0, len(repoIDs))
	for _, repoID := range repoIDs {
		recordCounts = append(recordCounts, recordCount{
			repoID:       repoID,
			uploadsCount: uploadsCounts[repoID],
			indexesCount: indexesCounts[repoID],
		})
	}

	return recordCounts
}
