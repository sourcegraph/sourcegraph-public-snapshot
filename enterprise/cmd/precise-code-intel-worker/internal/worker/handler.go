package worker

import (
	"compress/gzip"
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type handler struct {
	dbStore         DBStore
	lsifStore       LSIFStore
	uploadStore     uploadstore.Store
	gitserverClient GitserverClient
	metrics         metrics.WorkerMetrics
	enableBudget    bool
	budgetRemaining int64
}

var _ dbworker.Handler = &handler{}
var _ workerutil.WithPreDequeue = &handler{}
var _ workerutil.WithHooks = &handler{}

func (h *handler) Handle(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
	upload := record.(store.Upload)
	store := h.dbStore.With(tx)

	_, err := h.handle(ctx, store, upload)
	return err
}

func (h *handler) PreDequeue(ctx context.Context) (bool, interface{}, error) {
	if !h.enableBudget {
		return true, nil, nil
	}

	budgetRemaining := atomic.LoadInt64(&h.budgetRemaining)
	if budgetRemaining <= 0 {
		return false, nil, nil
	}

	return true, []*sqlf.Query{sqlf.Sprintf("(upload_size IS NULL OR upload_size <= %s)", budgetRemaining)}, nil
}

func (h *handler) PreHandle(ctx context.Context, record workerutil.Record) {
	atomic.AddInt64(&h.budgetRemaining, -h.getSize(record))
}

func (h *handler) PostHandle(ctx context.Context, record workerutil.Record) {
	atomic.AddInt64(&h.budgetRemaining, +h.getSize(record))
}

func (h *handler) getSize(record workerutil.Record) int64 {
	if size := record.(store.Upload).UploadSize; size != nil {
		return *size
	}

	return 0
}

// CloneInProgressDelay is the delay between processing attempts when a repo is currently being cloned.
const CloneInProgressDelay = time.Minute

// handle converts a raw upload into a dump within the given transaction context. Returns true if the
// upload record was requeued and false otherwise.
func (h *handler) handle(ctx context.Context, dbStore DBStore, upload store.Upload) (_ bool, err error) {
	// Ensure that the repo and revision are resolvable. If the repo does not exist, or if the repo has finished
	// cloning and the revision does not exist, then the upload will fail to process. If the repo is currently
	// cloning, then we'll requeue the upload to be tried again later. This will not increase the reset count
	// of the record (so this doesn't count against the upload as a legitimate attempt).
	if cloneInProgress, err := h.isRepoCurrentlyCloning(ctx, upload.RepositoryID, upload.Commit); err != nil {
		return false, err
	} else if cloneInProgress {
		if err := dbStore.Requeue(ctx, upload.ID, time.Now().UTC().Add(CloneInProgressDelay)); err != nil {
			return false, errors.Wrap(err, "store.Requeue")
		}

		return true, nil
	}

	uploadFilename := fmt.Sprintf("upload-%d.lsif.gz", upload.ID)

	// Pull raw uploaded data from bucket
	rc, err := h.uploadStore.Get(ctx, uploadFilename, 0)
	if err != nil {
		return false, errors.Wrap(err, "uploadStore.Get")
	}
	defer rc.Close()

	rc, err = gzip.NewReader(rc)
	if err != nil {
		return false, errors.Wrap(err, "gzip.NewReader")
	}
	defer rc.Close()

	defer func() {
		if err == nil {
			// Remove upload file after processing - we won't need it anymore. On failure we
			// may want to retry, so we should keep the upload data around for a bit. Older
			// uploads will be cleaned up periodically.
			if deleteErr := h.uploadStore.Delete(ctx, uploadFilename); deleteErr != nil {
				log15.Warn("Failed to delete upload file", "err", err)
			}
		}
	}()

	getChildren := func(ctx context.Context, dirnames []string) (map[string][]string, error) {
		directoryChildren, err := h.gitserverClient.DirectoryChildren(ctx, dbStore, upload.RepositoryID, upload.Commit, dirnames)
		if err != nil {
			return nil, errors.Wrap(err, "gitserverClient.DirectoryChildren")
		}
		return directoryChildren, nil
	}

	groupedBundleData, err := correlation.Correlate(ctx, rc, upload.ID, upload.Root, getChildren, h.metrics)
	if err != nil {
		return false, errors.Wrap(err, "correlation.Correlate")
	}

	if err := h.write(ctx, upload.ID, groupedBundleData); err != nil {
		return false, err
	}

	// Start a nested transaction. In the event that something after this point fails, we want to
	// update the upload record with an error message but do not want to alter any other data in
	// the database. Rolling back to this savepoint will allow us to discard any other changes
	// but still commit the transaction as a whole.

	// with Postgres savepoints. In the event that something after this point fails, we want to
	// update the upload record with an error message but do not want to alter any other data in
	// the database. Rolling back to this savepoint will allow us to discard any other changes
	// but still commit the transaction as a whole.
	tx, err := dbStore.Transact(ctx)
	if err != nil {
		return false, errors.Wrap(err, "store.Transact")
	}
	defer func() {
		err = tx.Done(err)
	}()

	if err := h.updateXrepoData(ctx, dbStore, upload, groupedBundleData.Packages, groupedBundleData.PackageReferences); err != nil {
		return false, err
	}

	return false, nil
}

// isRepoCurrentlyCloning determines if the target repository is currently being cloned.
// This function returns an error if the repo or commit cannot be resolved.
func (h *handler) isRepoCurrentlyCloning(ctx context.Context, repoID int, commit string) (_ bool, err error) {
	ctx, endOperation := h.metrics.RepoStateOperation.With(ctx, &err, observation.Args{})
	defer endOperation(1, observation.Args{})

	repo, err := backend.Repos.Get(ctx, api.RepoID(repoID))
	if err != nil {
		return false, errors.Wrap(err, "Repos.Get")
	}

	if _, err := backend.Repos.ResolveRev(ctx, repo, commit); err != nil {
		if vcs.IsCloneInProgress(err) {
			return true, nil
		}

		return false, errors.Wrap(err, "Repos.ResolveRev")
	}

	return false, nil
}

// write commits the correlated data to the database.
func (h *handler) write(ctx context.Context, id int, groupedBundleData *correlation.GroupedBundleData) (err error) {
	ctx, endOperation := h.metrics.WriteOperation.With(ctx, &err, observation.Args{})
	defer endOperation(1, observation.Args{})

	store, err := h.lsifStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = store.Done(err)
	}()

	if err := store.WriteMeta(ctx, id, groupedBundleData.Meta); err != nil {
		return errors.Wrap(err, "store.WriteMeta")
	}
	if err := store.WriteDocuments(ctx, id, groupedBundleData.Documents); err != nil {
		return errors.Wrap(err, "store.WriteDocuments")
	}
	if err := store.WriteResultChunks(ctx, id, groupedBundleData.ResultChunks); err != nil {
		return errors.Wrap(err, "writer.WriteResultChunks")
	}
	if err := store.WriteDefinitions(ctx, id, groupedBundleData.Definitions); err != nil {
		return errors.Wrap(err, "store.WriteDefinitions")
	}
	if err := store.WriteReferences(ctx, id, groupedBundleData.References); err != nil {
		return errors.Wrap(err, "store.WriteReferences")
	}

	return nil
}

// TODO(efritz) - refactor/simplify this after last change
func (h *handler) updateXrepoData(ctx context.Context, dbStore DBStore, upload store.Upload, packages []lsifstore.Package, packageReferences []lsifstore.PackageReference) (err error) {
	ctx, endOperation := h.metrics.UpdateXrepoDatabaseOperation.With(ctx, &err, observation.Args{})
	defer endOperation(1, observation.Args{})

	// Update package and package reference data to support cross-repo queries.
	if err := dbStore.UpdatePackages(ctx, packages); err != nil {
		return errors.Wrap(err, "store.UpdatePackages")
	}
	if err := dbStore.UpdatePackageReferences(ctx, packageReferences); err != nil {
		return errors.Wrap(err, "store.UpdatePackageReferences")
	}

	// Before we mark the upload as complete, we need to delete any existing completed uploads
	// that have the same repository_id, commit, root, and indexer values. Otherwise the transaction
	// will fail as these values form a unique constraint.
	if err := dbStore.DeleteOverlappingDumps(ctx, upload.RepositoryID, upload.Commit, upload.Root, upload.Indexer); err != nil {
		return errors.Wrap(err, "store.DeleteOverlappingDumps")
	}

	// Almost-success: we need to mark this upload as complete at this point as the next step changes	// the visibility of the dumps for this repository. This requires that the new dump be available in
	// the lsif_dumps view, which requires a change of state. In the event of a future failure we can
	// still roll back to the save point and mark the upload as errored.
	if err := dbStore.MarkComplete(ctx, upload.ID); err != nil {
		return errors.Wrap(err, "store.MarkComplete")
	}

	// Mark this repository so that the commit updater process will pull the full commit graph from gitserver
	// and recalculate the nearest upload for each commit as well as which uploads are visible from the tip of
	// the default branch. We don't do this inside of the transaction as we re-calcalute the entire set of data
	// from scratch and we want to be able to coalesce requests for the same repository rather than having a set
	// of uploads for the same repo re-calculate nearly identical data multiple times.
	if err := dbStore.MarkRepositoryAsDirty(ctx, upload.RepositoryID); err != nil {
		return errors.Wrap(err, "store.MarkRepositoryDirty")
	}

	return nil
}
