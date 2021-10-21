package worker

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/honeycombio/libhoney-go"
	"github.com/inconshreveable/log15"
	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type handler struct {
	dbStore         DBStore
	workerStore     dbworkerstore.Store
	lsifStore       LSIFStore
	uploadStore     uploadstore.Store
	gitserverClient GitserverClient
	enableBudget    bool
	budgetRemaining int64
}

var (
	_ workerutil.Handler        = &handler{}
	_ workerutil.WithPreDequeue = &handler{}
	_ workerutil.WithHooks      = &handler{}
)

// errCommitDoesNotExist occurs when gitserver does not recognize the commit attached to the upload.
var errCommitDoesNotExist = errors.Errorf("commit does not exist")

func (h *handler) Handle(ctx context.Context, record workerutil.Record) error {
	_, err := h.handle(ctx, record.(store.Upload))
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

// handle converts a raw upload into a dump within the given transaction context. Returns true if the
// upload record was requeued and false otherwise.
func (h *handler) handle(ctx context.Context, upload store.Upload) (requeued bool, err error) {
	start := time.Now()
	defer func() {
		if honey.Enabled() {
			_ = createHoneyEvent(ctx, upload, err, time.Since(start)).Send()
		}
	}()

	repo, err := backend.Repos.Get(ctx, api.RepoID(upload.RepositoryID))
	if err != nil {
		return false, errors.Wrap(err, "Repos.Get")
	}

	if requeued, err := requeueIfCloning(ctx, h.workerStore, upload, repo); err != nil || requeued {
		return requeued, err
	}

	// Determine if the upload is for the default Git branch.
	isDefaultBranch, err := h.gitserverClient.DefaultBranchContains(ctx, upload.RepositoryID, upload.Commit)
	if err != nil {
		return false, errors.Wrap(err, "gitserver.DefaultBranchContains")
	}

	getChildren := func(ctx context.Context, dirnames []string) (map[string][]string, error) {
		directoryChildren, err := h.gitserverClient.DirectoryChildren(ctx, upload.RepositoryID, upload.Commit, dirnames)
		if err != nil {
			return nil, errors.Wrap(err, "gitserverClient.DirectoryChildren")
		}
		return directoryChildren, nil
	}

	return false, withUploadData(ctx, h.uploadStore, upload.ID, func(r io.Reader) (err error) {
		groupedBundleData, err := conversion.Correlate(ctx, r, upload.Root, getChildren)
		if err != nil {
			return errors.Wrap(err, "conversion.Correlate")
		}

		// Note: this is writing to a different database than the block below, so we need to use a
		// different transaction context (managed by the writeData function).
		if err := writeData(ctx, h.lsifStore, upload, repo, isDefaultBranch, groupedBundleData); err != nil {
			if isUniqueConstraintViolation(err) {
				// If this is a unique constraint violation, then we've previously processed this same
				// upload record up to this point, but failed to perform the transaction below. We can
				// safely assume that the entire index's data is in the codeintel database, as it's
				// parsed determinstically and written atomically.
				log15.Warn("LSIF data already exists for upload record")
			} else {
				return err
			}
		}

		// Start a nested transaction with Postgres savepoints. In the event that something after this
		// point fails, we want to update the upload record with an error message but do not want to
		// alter any other data in the database. Rolling back to this savepoint will allow us to discard
		// any other changes but still commit the transaction as a whole.
		return inTransaction(ctx, h.dbStore, func(tx DBStore) error {
			// Find the date of the commit and store that in the upload record. We do this now as we
			// will need to find the _oldest_ commit with code intelligence data to efficiently update
			// the commit graph for the repository.
			_, commitDate, revisionExists, err := h.gitserverClient.CommitDate(ctx, upload.RepositoryID, upload.Commit)
			if err != nil {
				return errors.Wrap(err, "gitserverClient.CommitDate")
			}
			if !revisionExists {
				return errCommitDoesNotExist
			}
			if err := tx.UpdateCommitedAt(ctx, upload.ID, commitDate); err != nil {
				return errors.Wrap(err, "store.CommitDate")
			}

			// Update package and package reference data to support cross-repo queries.
			if err := tx.UpdatePackages(ctx, upload.ID, groupedBundleData.Packages); err != nil {
				return errors.Wrap(err, "store.UpdatePackages")
			}
			if err := tx.UpdatePackageReferences(ctx, upload.ID, groupedBundleData.PackageReferences); err != nil {
				return errors.Wrap(err, "store.UpdatePackageReferences")
			}

			// When inserting a new completed upload record, update the reference counts both to it from
			// existing uploads, as well as the reference counts to all of this new upload's dependencies.
			// We decrement reference counts of dependencies on upload deletion, so this count should
			// always be up to date as records are created and removed.
			if err := tx.UpdateNumReferences(ctx, []int{upload.ID}); err != nil {
				return errors.Wrap(err, "store.UpdateNumReferences")
			}
			if err := tx.UpdateDependencyNumReferences(ctx, []int{upload.ID}, false); err != nil {
				return errors.Wrap(err, "store.UpdateDependencyNumReferences")
			}

			// Before we mark the upload as complete, we need to delete any existing completed uploads
			// that have the same repository_id, commit, root, and indexer values. Otherwise the transaction
			// will fail as these values form a unique constraint.
			if err := tx.DeleteOverlappingDumps(ctx, upload.RepositoryID, upload.Commit, upload.Root, upload.Indexer); err != nil {
				return errors.Wrap(err, "store.DeleteOverlappingDumps")
			}

			// Insert a companion record to this upload that will asynchronously trigger other workers to
			// sync/create referenced dependency repositories and queue auto-index records for the monikers
			// written into the lsif_references table attached by this index processing job.
			if _, err := tx.InsertDependencySyncingJob(ctx, upload.ID); err != nil {
				return errors.Wrap(err, "store.InsertDependencyIndexingJob")
			}

			// Mark this repository so that the commit updater process will pull the full commit graph from
			// gitserver and recalculate the nearest upload for each commit as well as which uploads are visible
			// from the tip of the default branch. We don't do this inside of the transaction as we re-calcalute
			// the entire set of data from scratch and we want to be able to coalesce requests for the same
			// repository rather than having a set of uploads for the same repo re-calculate nearly identical
			// data multiple times.
			if err := tx.MarkRepositoryAsDirty(ctx, upload.RepositoryID); err != nil {
				return errors.Wrap(err, "store.MarkRepositoryAsDirty")
			}

			return nil
		})
	})
}

func inTransaction(ctx context.Context, dbStore DBStore, fn func(tx DBStore) error) (err error) {
	tx, err := dbStore.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "store.Transact")
	}
	defer func() { err = tx.Done(err) }()

	return fn(tx)
}

// CloneInProgressDelay is the delay between processing attempts when a repo is currently being cloned.
const CloneInProgressDelay = time.Minute

// requeueIfCloning ensures that the repo and revision are resolvable. If the repo does not exist, or
// if the repo has finished cloning and the revision does not exist, then the upload will fail to process.
// If the repo is currently cloning, then we'll requeue the upload to be tried again later. This will not
// increase the reset count of the record (so this doesn't count against the upload as a legitimate attempt).
func requeueIfCloning(ctx context.Context, workerStore dbworkerstore.Store, upload store.Upload, repo *types.Repo) (requeued bool, _ error) {
	if _, err := backend.Repos.ResolveRev(ctx, repo, upload.Commit); err != nil {
		if !gitdomain.IsCloneInProgress(err) {
			return false, errors.Wrap(err, "Repos.ResolveRev")
		}

		if err := workerStore.Requeue(ctx, upload.ID, time.Now().UTC().Add(CloneInProgressDelay)); err != nil {
			return false, errors.Wrap(err, "store.Requeue")
		}

		return true, nil
	}

	return false, nil
}

// withUploadData will invoke the given function with a reader of the upload's raw data. The
// consumer should expect raw newline-delimited JSON content. If the function returns without
// an error, the upload file will be deleted.
func withUploadData(ctx context.Context, uploadStore uploadstore.Store, id int, fn func(r io.Reader) error) error {
	uploadFilename := fmt.Sprintf("upload-%d.lsif.gz", id)

	// Pull raw uploaded data from bucket
	rc, err := uploadStore.Get(ctx, uploadFilename)
	if err != nil {
		return errors.Wrap(err, "uploadStore.Get")
	}
	defer rc.Close()

	rc, err = gzip.NewReader(rc)
	if err != nil {
		return errors.Wrap(err, "gzip.NewReader")
	}
	defer rc.Close()

	if err := fn(rc); err != nil {
		return err
	}

	if err := uploadStore.Delete(ctx, uploadFilename); err != nil {
		log15.Warn("Failed to delete upload file", "err", err, "filename", uploadFilename)
	}

	return nil
}

// writeData transactionally writes the given grouped bundle data into the given LSIF store.
func writeData(ctx context.Context, lsifStore LSIFStore, upload dbstore.Upload, repo *types.Repo, isDefaultBranch bool, groupedBundleData *precise.GroupedBundleDataChans) (err error) {
	// Upsert values used for documentation search that have high contention. We do this with the raw LSIF store
	// instead of in the transaction below because the rows being upserted tend to have heavy contention.
	repositoryNameID, languageNameID, err := lsifStore.WriteDocumentationSearchPrework(ctx, upload, repo, isDefaultBranch)
	if err != nil {
		return errors.Wrap(err, "store.WriteDocumentationSearchPrework")
	}

	tx, err := lsifStore.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.WriteMeta(ctx, upload.ID, groupedBundleData.Meta); err != nil {
		return errors.Wrap(err, "store.WriteMeta")
	}
	if err := tx.WriteDocuments(ctx, upload.ID, groupedBundleData.Documents); err != nil {
		return errors.Wrap(err, "store.WriteDocuments")
	}
	if err := tx.WriteResultChunks(ctx, upload.ID, groupedBundleData.ResultChunks); err != nil {
		return errors.Wrap(err, "store.WriteResultChunks")
	}
	if err := tx.WriteDefinitions(ctx, upload.ID, groupedBundleData.Definitions); err != nil {
		return errors.Wrap(err, "store.WriteDefinitions")
	}
	if err := tx.WriteReferences(ctx, upload.ID, groupedBundleData.References); err != nil {
		return errors.Wrap(err, "store.WriteReferences")
	}
	if err := tx.WriteDocumentationPages(ctx, upload, repo, isDefaultBranch, groupedBundleData.DocumentationPages, repositoryNameID, languageNameID); err != nil {
		return errors.Wrap(err, "store.WriteDocumentationPages")
	}
	if err := tx.WriteDocumentationPathInfo(ctx, upload.ID, groupedBundleData.DocumentationPathInfo); err != nil {
		return errors.Wrap(err, "store.WriteDocumentationPathInfo")
	}
	if err := tx.WriteDocumentationMappings(ctx, upload.ID, groupedBundleData.DocumentationMappings); err != nil {
		return errors.Wrap(err, "store.WriteDocumentationMappings")
	}

	return nil
}

func isUniqueConstraintViolation(err error) bool {
	var e *pgconn.PgError
	return errors.As(err, &e) && e.Code == "23505"
}

func createHoneyEvent(ctx context.Context, upload store.Upload, err error, duration time.Duration) *libhoney.Event {
	fields := map[string]interface{}{
		"duration_ms":    duration.Milliseconds(),
		"uploadID":       upload.ID,
		"repositoryID":   upload.RepositoryID,
		"repositoryName": upload.RepositoryName,
		"commit":         upload.Commit,
		"root":           upload.Root,
		"indexer":        upload.Indexer,
	}

	if err != nil {
		fields["error"] = err.Error()
	}
	if upload.UploadSize != nil {
		fields["uploadSize"] = upload.UploadSize
	}
	if traceID := trace.ID(ctx); traceID != "" {
		fields["trace"] = trace.URL(traceID)
		fields["traceID"] = traceID
	}

	return honey.EventWithFields("codeintel-worker", fields)
}
