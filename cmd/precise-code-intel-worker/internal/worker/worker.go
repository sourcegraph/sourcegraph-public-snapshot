package worker

import (
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/existence"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/writer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

type WorkerOpts struct {
	DB                  db.DB
	BundleManagerClient bundles.BundleManagerClient
	GitserverClient     gitserver.Client
	PollInterval        time.Duration
}

type Worker struct {
	db                  db.DB
	bundleManagerClient bundles.BundleManagerClient
	gitserverClient     gitserver.Client
	pollInterval        time.Duration
}

func New(opts WorkerOpts) *Worker {
	return &Worker{
		db:                  opts.DB,
		bundleManagerClient: opts.BundleManagerClient,
		gitserverClient:     opts.GitserverClient,
		pollInterval:        opts.PollInterval,
	}
}

func (w *Worker) Start() error {
	for {
		if ok, err := w.dequeueAndProcess(); err != nil {
			return err
		} else if !ok {
			time.Sleep(w.pollInterval)
		}
	}
}

// dequeueAndProcess pulls a job from the queue and processes it. If there was no job ready
// to process, this method returns a false-valued flag. Only critical errors are returned.
// Processing errors are only written to the upload record and are not expected to be handled
// by the calling function.
func (w *Worker) dequeueAndProcess() (_ bool, err error) {
	upload, jobHandle, ok, err := w.db.Dequeue(context.Background())
	if err != nil || !ok {
		return false, err
	}
	defer func() {
		if closeErr := jobHandle.CloseTx(err); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	if err = process(context.Background(), w.db, w.bundleManagerClient, w.gitserverClient, upload, jobHandle); err != nil {
		log15.Warn("Failed to process upload", "id", upload.ID, "err", err)

		if markErr := jobHandle.MarkErrored(err.Error(), ""); markErr != nil {
			return false, markErr
		}
	}

	return true, nil
}

// process converts a raw upload into a dump within the given job handle context.
func process(
	ctx context.Context,
	db db.DB,
	bundleManagerClient bundles.BundleManagerClient,
	gitserverClient gitserver.Client,
	upload db.Upload,
	jobHandle db.JobHandle,
) (err error) {
	// Create scratch directory that we can clean on completion/failure
	name, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer func() {
		if cleanupErr := os.RemoveAll(name); cleanupErr != nil {
			log15.Warn("Failed to remove temporary directory", "path", name, "err", cleanupErr)
		}
	}()

	// Pull raw uploaded data from bundle manager
	filename, err := bundleManagerClient.GetUpload(ctx, upload.ID, name)
	if err != nil {
		return err
	}

	// Create target file for converted database
	uuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	newFilename := filepath.Join(name, uuid.String())

	// Read raw upload and write converted database to newFilename. This process also correlates
	// and returns the  data we need to insert into Postgres to support cross-dump/repo queries.
	packages, packageReferences, err := convert(
		context.Background(),
		filename,
		newFilename,
		upload.ID,
		upload.Root,
		func(dirnames []string) (map[string][]string, error) {
			return gitserverClient.DirectoryChildren(db, upload.RepositoryID, upload.Commit, dirnames)
		},
	)
	if err != nil {
		return err
	}

	// At this point we haven't touched the database. We're going to start a nested transaction
	// with Postgres savepoints. In the event that something after this point fails, we want to
	// update the upload record with an error message but do not want to alter any other data in
	// the database. Rolling back to this savepoint will allow us to discard any other changes
	// but still commit the transaction as a whole.
	if err := jobHandle.Savepoint(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rollbackErr := jobHandle.RollbackToLastSavepoint(); rollbackErr != nil {
				err = multierror.Append(err, rollbackErr)
			}
		}
	}()

	// Update package and package reference data to support cross-repo queries.
	if err := db.UpdatePackages(context.Background(), jobHandle.Tx(), packages); err != nil {
		return err
	}
	if err := db.UpdatePackageReferences(context.Background(), jobHandle.Tx(), packageReferences); err != nil {
		return err
	}

	// Before we mark the upload as complete, we need to delete any existing completed uploads
	// that have the same repository_id, commit, root, and indexer values. Otherwise the transaction
	// will fail as these values form a unique constraint.
	if err := db.DeleteOverlappingDumps(ctx, jobHandle.Tx(), upload.RepositoryID, upload.Commit, upload.Root, upload.Indexer); err != nil {
		return err
	}

	// Almost-success: we need to mark this upload as complete at this point as the next step changes
	// the visibility of the dumps for this repository. This requires that the new dump be available in
	// the lsif_dumps view, which requires a change of state. In the event of a future failure we can
	// still roll back to the save point and mark the upload as errored.
	if err := jobHandle.MarkComplete(); err != nil {
		return err
	}

	// Discover commits around the current tip commit and the commit of this upload. Upsert these
	// commits into the lsif_commits table, then update the visibility of all dumps for this repository.
	if err := updateCommitsAndVisibility(ctx, db, gitserverClient, jobHandle.Tx(), upload.RepositoryID, upload.Commit); err != nil {
		return err
	}

	f, err := os.Open(newFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Send converted database file to bundle manager
	if err := bundleManagerClient.SendDB(ctx, upload.ID, f); err != nil {
		return err
	}

	return nil
}

// updateCommits updates the lsif_commits table with the current data known to gitserver, then updates the
// visibility of all dumps for the given repository.
func updateCommitsAndVisibility(ctx context.Context, db db.DB, gitserverClient gitserver.Client, tx *sql.Tx, repositoryID int, commit string) error {
	tipCommit, err := gitserverClient.Head(db, repositoryID)
	if err != nil {
		return err
	}
	newCommits, err := gitserverClient.CommitsNear(db, repositoryID, tipCommit)
	if err != nil {
		return err
	}

	if tipCommit != commit {
		// If the tip is ahead of this commit, we also want to discover all of the commits between this
		// commit and the tip so that we can accurately determine what is visible from the tip. If we
		// do not do this before the updateDumpsVisibleFromTip call below, no dumps will be reachable
		// from the tip and all dumps will be invisible.
		additionalCommits, err := gitserverClient.CommitsNear(db, repositoryID, commit)
		if err != nil {
			return err
		}

		for k, vs := range additionalCommits {
			newCommits[k] = append(newCommits[k], vs...)
		}
	}

	// TODO - need to do same discover on query
	// TODO - determine if we know about these commits first
	if err := db.UpdateCommits(ctx, tx, repositoryID, newCommits); err != nil {
		return err
	}

	if err := db.UpdateDumpsVisibleFromTip(ctx, tx, repositoryID, tipCommit); err != nil {
		return err
	}

	return nil
}

// convert correlates the raw input data and commits the correlated data to disk.
func convert(
	ctx context.Context,
	filename string,
	newFilename string,
	dumpID int,
	root string,
	getChildren existence.GetChildrenFunc,
) (_ []types.Package, _ []types.PackageReference, err error) {
	groupedBundleData, err := correlation.Correlate(filename, dumpID, root, getChildren)
	if err != nil {
		return nil, nil, err
	}

	if err := write(ctx, newFilename, groupedBundleData); err != nil {
		return nil, nil, err
	}

	return groupedBundleData.Packages, groupedBundleData.PackageReferences, nil
}

// write commits the correlated data to disk.
func write(ctx context.Context, filename string, groupedBundleData *correlation.GroupedBundleData) error {
	writer, err := writer.NewSQLiteWriter(filename, serializer.NewDefaultSerializer())
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	fns := []func() error{
		func() error {
			return writer.WriteMeta(ctx, groupedBundleData.LSIFVersion, groupedBundleData.NumResultChunks)
		},
		func() error { return writer.WriteDocuments(ctx, groupedBundleData.Documents) },
		func() error { return writer.WriteResultChunks(ctx, groupedBundleData.ResultChunks) },
		func() error { return writer.WriteDefinitions(ctx, groupedBundleData.Definitions) },
		func() error { return writer.WriteReferences(ctx, groupedBundleData.References) },
		func() error { return writer.Flush(ctx) },
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}
