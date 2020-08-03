package worker

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/correlation"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	sqlitewriter "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// CloneInProgressDelay is the delay between processing attempts when a repo is currently being cloned.
const CloneInProgressDelay = time.Minute

// Processor converts raw uploads into dumps.
type Processor interface {
	Process(ctx context.Context, tx store.Store, upload store.Upload) (bool, error)
}

type processor struct {
	bundleManagerClient bundles.BundleManagerClient
	gitserverClient     gitserver.Client
	metrics             metrics.WorkerMetrics
}

// process converts a raw upload into a dump within the given transaction context. Returns true if the
// upload record was requeued and false otherwise.
func (p *processor) Process(ctx context.Context, store store.Store, upload store.Upload) (_ bool, err error) {
	// Ensure that the repo and revision are resolvable. If the repo does not exist, or if the repo has finished
	// cloning and the revision does not exist, then the upload will fail to process. If the repo is currently
	// cloning, then we'll requeue the upload to be tried again later. This will not increase the reset count
	// of the record (so this doesn't count against the upload as a legitimate attempt).
	if cloneInProgress, err := p.isRepoCurrentlyCloning(ctx, upload.RepositoryID, upload.Commit); err != nil {
		return false, err
	} else if cloneInProgress {
		if err := store.Requeue(ctx, upload.ID, time.Now().UTC().Add(CloneInProgressDelay)); err != nil {
			return false, errors.Wrap(err, "store.Requeue")
		}

		return true, nil
	}

	// Create scratch directory that we can clean on completion/failure
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return false, err
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			log15.Warn("Failed to remove temporary directory", "path", tempDir, "err", cleanupErr)
		}
	}()

	// Pull raw uploaded data from bundle manager
	r, err := p.bundleManagerClient.GetUpload(ctx, upload.ID)
	if err != nil {
		return false, errors.Wrap(err, "bundleManager.GetUpload")
	}
	defer func() {
		if err != nil {
			// Remove upload file on error instead of waiting for it to expire
			if deleteErr := p.bundleManagerClient.DeleteUpload(ctx, upload.ID); deleteErr != nil {
				log15.Warn("Failed to delete upload file", "err", err)
			}
		}
	}()

	getChildren := func(ctx context.Context, dirnames []string) (map[string][]string, error) {
		directoryChildren, err := p.gitserverClient.DirectoryChildren(ctx, store, upload.RepositoryID, upload.Commit, dirnames)
		if err != nil {
			return nil, errors.Wrap(err, "gitserverClient.DirectoryChildren")
		}
		return directoryChildren, nil
	}

	groupedBundleData, err := correlation.Correlate(ctx, r, upload.ID, upload.Root, getChildren, p.metrics)
	if err != nil {
		return false, errors.Wrap(err, "correlation.Correlate")
	}

	if err := p.write(ctx, tempDir, groupedBundleData); err != nil {
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
	tx, err := store.Transact(ctx)
	if err != nil {
		return false, errors.Wrap(err, "store.Transact")
	}
	defer func() {
		err = tx.Done(err)
	}()

	if err := p.updateXrepoData(ctx, store, upload, groupedBundleData.Packages, groupedBundleData.PackageReferences); err != nil {
		return false, err
	}

	// Send converted database file to bundle manager
	if err := p.sendDB(ctx, upload.ID, filepath.Join(tempDir, "sqlite.db")); err != nil {
		return false, err
	}

	return false, nil
}

// isRepoCurrentlyCloning determines if the target repository is currently being cloned.
// This function returns an error if the repo or commit cannot be resolved.
func (p *processor) isRepoCurrentlyCloning(ctx context.Context, repoID int, commit string) (_ bool, err error) {
	ctx, endOperation := p.metrics.RepoStateOperation.With(ctx, &err, observation.Args{})
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

// write commits the correlated data to disk.
func (p *processor) write(ctx context.Context, dirname string, groupedBundleData *correlation.GroupedBundleData) (err error) {
	ctx, endOperation := p.metrics.WriteOperation.With(ctx, &err, observation.Args{})
	defer endOperation(1, observation.Args{})

	writer, err := sqlitewriter.NewWriter(ctx, filepath.Join(dirname, "sqlite.db"))
	if err != nil {
		return err
	}
	defer func() {
		err = writer.Close(err)
	}()

	if err := writer.WriteMeta(ctx, groupedBundleData.Meta); err != nil {
		return errors.Wrap(err, "writer.WriteMeta")
	}
	if err := writer.WriteDocuments(ctx, groupedBundleData.Documents); err != nil {
		return errors.Wrap(err, "writer.WriteDocuments")
	}
	if err := writer.WriteResultChunks(ctx, groupedBundleData.ResultChunks); err != nil {
		return errors.Wrap(err, "writer.WriteResultChunks")
	}
	if err := writer.WriteDefinitions(ctx, groupedBundleData.Definitions); err != nil {
		return errors.Wrap(err, "writer.WriteDefinitions")
	}
	if err := writer.WriteReferences(ctx, groupedBundleData.References); err != nil {
		return errors.Wrap(err, "writer.WriteReferences")
	}

	return err
}

// TODO(efritz) - refactor/simplify this after last change
func (p *processor) updateXrepoData(ctx context.Context, store store.Store, upload store.Upload, packages []types.Package, packageReferences []types.PackageReference) (err error) {
	ctx, endOperation := p.metrics.UpdateXrepoDatabaseOperation.With(ctx, &err, observation.Args{})
	defer endOperation(1, observation.Args{})

	// Update package and package reference data to support cross-repo queries.
	if err := store.UpdatePackages(ctx, packages); err != nil {
		return errors.Wrap(err, "store.UpdatePackages")
	}
	if err := store.UpdatePackageReferences(ctx, packageReferences); err != nil {
		return errors.Wrap(err, "store.UpdatePackageReferences")
	}

	// Before we mark the upload as complete, we need to delete any existing completed uploads
	// that have the same repository_id, commit, root, and indexer values. Otherwise the transaction
	// will fail as these values form a unique constraint.
	if err := store.DeleteOverlappingDumps(ctx, upload.RepositoryID, upload.Commit, upload.Root, upload.Indexer); err != nil {
		return errors.Wrap(err, "store.DeleteOverlappingDumps")
	}

	// Almost-success: we need to mark this upload as complete at this point as the next step changes
	// the visibility of the dumps for this repository. This requires that the new dump be available in
	// the lsif_dumps view, which requires a change of state. In the event of a future failure we can
	// still roll back to the save point and mark the upload as errored.
	if err := store.MarkComplete(ctx, upload.ID); err != nil {
		return errors.Wrap(err, "store.MarkComplete")
	}

	// Mark this repository so that the commit updater process will pull the full commit graph from gitserver
	// and recalculate the nearest upload for each commit as well as which uploads are visible from the tip of
	// the default branch. We don't do this inside of the transaction as we re-calcalute the entire set of data
	// from scratch and we want to be able to coalesce requests for the same repository rather than having a set
	// of uploads for the same repo re-calculate nearly identical data multiple times.
	if err := store.MarkRepositoryAsDirty(ctx, upload.RepositoryID); err != nil {
		return errors.Wrap(err, "store.MarkRepositoryDirty")
	}

	return nil
}

func (p *processor) sendDB(ctx context.Context, uploadID int, tempDir string) (err error) {
	ctx, endOperation := p.metrics.SendDBOperation.With(ctx, &err, observation.Args{})
	defer endOperation(1, observation.Args{})

	if err := p.bundleManagerClient.SendDB(ctx, uploadID, tempDir); err != nil {
		return errors.Wrap(err, "bundleManager.SendDB")
	}

	return nil
}
