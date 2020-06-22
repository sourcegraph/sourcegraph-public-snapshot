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

// removeProcessedRecordsWithoutBundleFile removes all upload records in the
// processed state that do not have a corresponding bundle file on disk.
func (j *Janitor) removeProcessedRecordsWithoutBundleFile() error {
	ctx := context.Background()

	// TODO(efritz) - request in batches
	uploads, _, err := j.store.GetUploads(ctx, store.GetUploadsOptions{
		State: "processed",
	})
	if err != nil {
		return errors.Wrap(err, "store.GetUploads")
	}

	for _, upload := range uploads {
		exists, err := paths.PathExists(paths.DBDir(j.bundleDir, int64(upload.ID)))
		if err != nil {
			return errors.Wrap(err, "paths.PathExists")
		}
		if exists {
			continue
		}

		deleted, err := j.store.DeleteUploadByID(ctx, upload.ID, j.getTipCommit)
		if err != nil {
			return errors.Wrap(err, "store.DeleteUploadByID")
		}

		if deleted {
			log15.Debug("Removed upload record with no bundle file", "id", upload.ID)
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

	// TODO(efritz) - request in batches
	uploads, _, err := j.store.GetUploads(ctx, store.GetUploadsOptions{
		State:          "uploading",
		UploadedBefore: &t,
	})
	if err != nil {
		return errors.Wrap(err, "store.GetUploads")
	}

	for _, upload := range uploads {
		deleted, err := j.store.DeleteUploadByID(ctx, upload.ID, j.getTipCommit)
		if err != nil {
			return errors.Wrap(err, "store.DeleteUploadByID")
		}

		if deleted {
			log15.Debug("Removed upload record stuck uploading", "id", upload.ID)
			j.metrics.UploadRecordsRemoved.Inc()
		}
	}

	return nil

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
