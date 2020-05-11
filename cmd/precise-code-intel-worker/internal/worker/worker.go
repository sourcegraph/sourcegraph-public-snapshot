package worker

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/existence"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/writer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

type Worker struct {
	db           db.DB
	processor    Processor
	pollInterval time.Duration
	metrics      WorkerMetrics
	done         chan struct{}
	once         sync.Once
}

func NewWorker(
	db db.DB,
	bundleManagerClient bundles.BundleManagerClient,
	gitserverClient gitserver.Client,
	pollInterval time.Duration,
	metrics WorkerMetrics,
) *Worker {
	processor := &processor{
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
	}

	return &Worker{
		db:           db,
		processor:    processor,
		pollInterval: pollInterval,
		metrics:      metrics,
		done:         make(chan struct{}),
	}
}

func (w *Worker) Start() {
	for {
		if ok, _ := w.dequeueAndProcess(context.Background()); !ok {
			select {
			case <-time.After(w.pollInterval):
			case <-w.done:
				return
			}
		} else {
			select {
			case <-w.done:
				return
			default:
			}
		}
	}
}

func (w *Worker) Stop() {
	w.once.Do(func() {
		close(w.done)
	})
}

// TODO(efritz) - use cancellable context

// dequeueAndProcess pulls a job from the queue and processes it. If there
// were no jobs ready to process, this method returns a false-valued flag.
func (w *Worker) dequeueAndProcess(ctx context.Context) (_ bool, err error) {
	start := time.Now()

	upload, jobHandle, ok, err := w.db.Dequeue(ctx)
	if err != nil || !ok {
		return false, errors.Wrap(err, "db.Dequeue")
	}
	defer func() {
		if closeErr := jobHandle.Done(err); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}

		// TODO(efritz) - set error if correlation failed
		w.metrics.Processor.Observe(time.Since(start).Seconds(), 1, &err)
	}()

	log15.Info("Dequeued upload for processing", "id", upload.ID)

	if processErr := w.processor.Process(ctx, jobHandle.DB(), upload, jobHandle); processErr == nil {
		log15.Info("Processed upload", "id", upload.ID)
	} else {
		// TODO(efritz) - distinguish between correlation and system errors
		log15.Warn("Failed to process upload", "id", upload.ID, "err", processErr)

		if markErr := jobHandle.MarkErrored(ctx, processErr.Error(), ""); markErr != nil {
			return true, errors.Wrap(markErr, "jobHandle.MarkErrored")
		}
	}

	return true, nil
}

// Processor converts raw uploads into dumps.
type Processor interface {
	Process(ctx context.Context, db db.DB, upload db.Upload, jobHandle db.JobHandle) error
}

type processor struct {
	bundleManagerClient bundles.BundleManagerClient
	gitserverClient     gitserver.Client
}

// process converts a raw upload into a dump within the given job handle context.
func (p *processor) Process(
	ctx context.Context,
	db db.DB,
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
	filename, err := p.bundleManagerClient.GetUpload(ctx, upload.ID, name)
	if err != nil {
		return errors.Wrap(err, "bundleManager.GetUpload")
	}
	defer func() {
		if err != nil {
			// Remove upload file on error instead of waiting for it to expire
			if deleteErr := p.bundleManagerClient.DeleteUpload(ctx, upload.ID); deleteErr != nil {
				log15.Warn("Failed to delete upload file", "err", err)
			}
		}
	}()

	// Create target file for converted database
	uuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	newFilename := filepath.Join(name, uuid.String())

	// Read raw upload and write converted database to newFilename. This process also correlates
	// and returns the  data we need to insert into Postgres to support cross-dump/repo queries.
	packages, packageReferences, err := convert(
		ctx,
		filename,
		newFilename,
		upload.ID,
		upload.Root,
		func(dirnames []string) (map[string][]string, error) {
			directoryChildren, err := p.gitserverClient.DirectoryChildren(db, upload.RepositoryID, upload.Commit, dirnames)
			if err != nil {
				return nil, errors.Wrap(err, "gitserverClient.DirectoryChildren")
			}
			return directoryChildren, nil
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
	if err := jobHandle.Savepoint(ctx); err != nil {
		return errors.Wrap(err, "jobHandle.Savepoint")
	}
	defer func() {
		if err != nil {
			if rollbackErr := jobHandle.RollbackToLastSavepoint(ctx); rollbackErr != nil {
				err = multierror.Append(err, rollbackErr)
			}
		}
	}()

	// Update package and package reference data to support cross-repo queries.
	if err := db.UpdatePackages(ctx, packages); err != nil {
		return errors.Wrap(err, "db.UpdatePackages")
	}
	if err := db.UpdatePackageReferences(ctx, packageReferences); err != nil {
		return errors.Wrap(err, "db.UpdatePackageReferences")
	}

	// Before we mark the upload as complete, we need to delete any existing completed uploads
	// that have the same repository_id, commit, root, and indexer values. Otherwise the transaction
	// will fail as these values form a unique constraint.
	if err := db.DeleteOverlappingDumps(ctx, upload.RepositoryID, upload.Commit, upload.Root, upload.Indexer); err != nil {
		return errors.Wrap(err, "db.DeleteOverlappingDumps")
	}

	// Almost-success: we need to mark this upload as complete at this point as the next step changes
	// the visibility of the dumps for this repository. This requires that the new dump be available in
	// the lsif_dumps view, which requires a change of state. In the event of a future failure we can
	// still roll back to the save point and mark the upload as errored.
	if err := jobHandle.MarkComplete(ctx); err != nil {
		return errors.Wrap(err, "jobHandle.MarkComplete")
	}

	// Discover commits around the current tip commit and the commit of this upload. Upsert these
	// commits into the lsif_commits table, then update the visibility of all dumps for this repository.
	if err := p.updateCommitsAndVisibility(ctx, db, upload.RepositoryID, upload.Commit); err != nil {
		return errors.Wrap(err, "updateCommitsAndVisibility")
	}

	f, err := os.Open(newFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Send converted database file to bundle manager
	if err := p.bundleManagerClient.SendDB(ctx, upload.ID, f); err != nil {
		return errors.Wrap(err, "bundleManager.SendDB")
	}

	return nil
}

// updateCommits updates the lsif_commits table with the current data known to gitserver, then updates the
// visibility of all dumps for the given repository.
func (p *processor) updateCommitsAndVisibility(ctx context.Context, db db.DB, repositoryID int, commit string) error {
	tipCommit, err := p.gitserverClient.Head(db, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}
	newCommits, err := p.gitserverClient.CommitsNear(db, repositoryID, tipCommit)
	if err != nil {
		return errors.Wrap(err, "gitserver.CommitsNear")
	}

	if tipCommit != commit {
		// If the tip is ahead of this commit, we also want to discover all of the commits between this
		// commit and the tip so that we can accurately determine what is visible from the tip. If we
		// do not do this before the updateDumpsVisibleFromTip call below, no dumps will be reachable
		// from the tip and all dumps will be invisible.
		additionalCommits, err := p.gitserverClient.CommitsNear(db, repositoryID, commit)
		if err != nil {
			return errors.Wrap(err, "gitserver.CommitsNear")
		}

		for k, vs := range additionalCommits {
			newCommits[k] = append(newCommits[k], vs...)
		}
	}

	if err := db.UpdateCommits(ctx, repositoryID, newCommits); err != nil {
		return errors.Wrap(err, "db.UpdateCommits")
	}

	if err := db.UpdateDumpsVisibleFromTip(ctx, repositoryID, tipCommit); err != nil {
		return errors.Wrap(err, "db.UpdateDumpsVisibleFromTip")
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
		return nil, nil, errors.Wrap(err, "correlation.Correlate")
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

	if err := writer.WriteMeta(ctx, groupedBundleData.LSIFVersion, groupedBundleData.NumResultChunks); err != nil {
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

	if err := writer.Flush(ctx); err != nil {
		return errors.Wrap(err, "writer.Flush")
	}

	return nil
}
