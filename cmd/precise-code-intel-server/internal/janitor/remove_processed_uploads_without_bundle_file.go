package janitor

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

// BundleBatchSize is the maximum number of bundle ids to request at
// once from the precise-code-intel-bundle-manager.
const BundleBatchSize = 100

// removeProcessedUploadsWithoutBundleFile removes all processed upload records
// that do not have a corresponding bundle file on disk.
func (j *Janitor) removeProcessedUploadsWithoutBundleFile() error {
	ctx := context.Background()

	ids, err := j.db.GetDumpIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "db.GetDumpIDs")
	}

	allExists := map[int]bool{}
	for _, batch := range batchIntSlice(ids, BundleBatchSize) {
		exists, err := j.bundleManagerClient.Exists(ctx, batch)
		if err != nil {
			return errors.Wrap(err, "bundleManagerClient.Exists")
		}

		for k, v := range exists {
			allExists[k] = v
		}
	}

	for id, exists := range allExists {
		if exists {
			continue
		}

		deleted, err := j.db.DeleteUploadByID(ctx, id, func(repositoryID int) (string, error) {
			tipCommit, err := gitserver.Head(j.db, repositoryID)
			if err != nil {
				return "", errors.Wrap(err, "gitserver.Head")
			}
			return tipCommit, nil
		})
		if err != nil {
			return errors.Wrap(err, "db.DeleteUploadByID")
		}
		if !deleted {
			continue
		}

		log15.Debug("Removed upload record with no bundle file", "id", id)
		j.metrics.UploadRecordsRemoved.Add(1)
	}

	return nil
}

// batchIntSlice returns slices of s (in order) at most batchSize in length.
func batchIntSlice(s []int, batchSize int) [][]int {
	batches := [][]int{}
	for len(s) > batchSize {
		batches = append(batches, s[:batchSize])
		s = s[batchSize:]
	}

	if len(s) > 0 {
		batches = append(batches, s)
	}

	return batches
}
