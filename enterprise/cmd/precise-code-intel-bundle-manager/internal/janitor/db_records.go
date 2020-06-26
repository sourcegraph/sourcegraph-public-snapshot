package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// GetUploadsBatchSize is the maximum number of uploads to request from the database at once.
const GetUploadsBatchSize = 100

// removeCompletedRecordsWithoutBundleFile removes all upload records in the
// completed state that do not have a corresponding bundle file on disk.
func (j *Janitor) removeCompletedRecordsWithoutBundleFile() error {
	ctx := context.Background()

	ids, err := j.getUploadIDs(ctx, store.GetUploadsOptions{
		State: "completed",
	})
	if err != nil {
		return err
	}

	for _, id := range ids {
		exists, err := paths.PathExists(paths.DBDir(j.bundleDir, int64(id)))
		if err != nil {
			return errors.Wrap(err, "paths.PathExists")
		}
		if exists {
			continue
		}

		deleted, err := j.store.DeleteUploadByID(ctx, id, j.getTipCommit)
		if err != nil {
			return errors.Wrap(err, "store.DeleteUploadByID")
		}

		if deleted {
			log15.Debug("Removed upload record with no bundle file", "id", id)
			j.metrics.UploadRecordsRemoved.Inc()
		}
	}

	return nil
}

// removeOldUploadingRecords removes all upload records in the uploading state that
// are older than the max upload part age.
func (j *Janitor) removeOldUploadingRecords() error {
	ctx := context.Background()
	t := time.Now().UTC().Add(-j.maxUploadPartAge)

	ids, err := j.getUploadIDs(ctx, store.GetUploadsOptions{
		State:          "uploading",
		UploadedBefore: &t,
	})
	if err != nil {
		return err
	}

	for _, id := range ids {
		deleted, err := j.store.DeleteUploadByID(ctx, id, j.getTipCommit)
		if err != nil {
			return errors.Wrap(err, "store.DeleteUploadByID")
		}

		if deleted {
			log15.Debug("Removed upload record stuck uploading", "id", id)
			j.metrics.UploadRecordsRemoved.Inc()
		}
	}

	return nil
}

// getUploadIDs returns the identifiers of all uploads matching the given options.
func (j *Janitor) getUploadIDs(ctx context.Context, opts store.GetUploadsOptions) ([]int, error) {
	var ids []int

	for {
		opts.Limit = GetUploadsBatchSize
		opts.Offset = len(ids)

		uploads, totalCount, err := j.store.GetUploads(ctx, opts)
		if err != nil {
			return nil, errors.Wrap(err, "store.GetUploads")
		}

		for i := range uploads {
			ids = append(ids, uploads[i].ID)
		}

		if len(ids) >= totalCount {
			break
		}
	}

	return ids, nil
}

// getTipCommit returns the head of the default branch for the given repository. This
// is used to recalculate the set of visible dumps for a repository on dump deletion.
func (j *Janitor) getTipCommit(ctx context.Context, repositoryID int) (string, error) {
	tipCommit, err := gitserver.Head(ctx, j.store, repositoryID)
	if err != nil && !isRepoNotExist(err) {
		return "", errors.Wrap(err, "gitserver.Head")
	}

	return tipCommit, nil
}

func isRepoNotExist(err error) bool {
	for err != nil {
		if vcs.IsRepoNotExist(err) {
			return true
		}

		err = errors.Unwrap(err)
	}

	return false
}
