package cleanup

import (
	"context"
	"sort"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (j *janitor) HandleDeletedRepository(ctx context.Context) (err error) {
	uploadsCounts, err := j.uploadSvc.DeleteUploadsWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DeleteUploadsWithoutRepository")
	}

	indexesCounts, err := j.uploadSvc.DeleteIndexesWithoutRepository(ctx, time.Now())
	if err != nil {
		return errors.Wrap(err, "uploadSvc.DeleteIndexesWithoutRepository")
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

// func (j *janitor) HandleError(err error) {
// 	j.metrics.numErrors.Inc()
// 	log15.Error("Failed to delete codeintel records with a deleted repository", "error", err)
// }

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
