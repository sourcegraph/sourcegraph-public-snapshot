package worker

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/existence"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	sqlitewriter "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

// Processor converts raw uploads into dumps.
type Processor interface {
	Process(ctx context.Context, tx db.DB, upload db.Upload) error
}

type processor struct {
	bundleManagerClient bundles.BundleManagerClient
	gitserverClient     gitserver.Client
}

// process converts a raw upload into a dump within the given transaction context.
func (p *processor) Process(ctx context.Context, tx db.DB, upload db.Upload) (err error) {
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

	// Create target file for converted database
	uuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	newFilename := filepath.Join(name, uuid.String())

	// Pull raw uploaded data from bundle manager
	r, err := p.bundleManagerClient.GetUpload(ctx, upload.ID)
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

	// Read raw upload and write converted database to newFilename. This process also correlates
	// and returns the  data we need to insert into Postgres to support cross-dump/repo queries.
	packages, packageReferences, err := convert(
		ctx,
		r,
		newFilename,
		upload.ID,
		upload.Root,
		func(dirnames []string) (map[string][]string, error) {
			directoryChildren, err := p.gitserverClient.DirectoryChildren(ctx, tx, upload.RepositoryID, upload.Commit, dirnames)
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
	savepointID, err := tx.Savepoint(ctx)
	if err != nil {
		return errors.Wrap(err, "db.Savepoint")
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.RollbackToSavepoint(ctx, savepointID); rollbackErr != nil {
				err = multierror.Append(err, rollbackErr)
			}
		}
	}()

	// Update package and package reference data to support cross-repo queries.
	if err := tx.UpdatePackages(ctx, packages); err != nil {
		return errors.Wrap(err, "db.UpdatePackages")
	}
	if err := tx.UpdatePackageReferences(ctx, packageReferences); err != nil {
		return errors.Wrap(err, "db.UpdatePackageReferences")
	}

	// Before we mark the upload as complete, we need to delete any existing completed uploads
	// that have the same repository_id, commit, root, and indexer values. Otherwise the transaction
	// will fail as these values form a unique constraint.
	if err := tx.DeleteOverlappingDumps(ctx, upload.RepositoryID, upload.Commit, upload.Root, upload.Indexer); err != nil {
		return errors.Wrap(err, "db.DeleteOverlappingDumps")
	}

	// Almost-success: we need to mark this upload as complete at this point as the next step changes
	// the visibility of the dumps for this repository. This requires that the new dump be available in
	// the lsif_dumps view, which requires a change of state. In the event of a future failure we can
	// still roll back to the save point and mark the upload as errored.
	if err := tx.MarkComplete(ctx, upload.ID); err != nil {
		return errors.Wrap(err, "db.MarkComplete")
	}

	// Discover commits around the current tip commit and the commit of this upload. Upsert these
	// commits into the lsif_commits table, then update the visibility of all dumps for this repository.
	if err := p.updateCommitsAndVisibility(ctx, tx, upload.RepositoryID, upload.Commit); err != nil {
		return errors.Wrap(err, "updateCommitsAndVisibility")
	}

	// Send converted database file to bundle manager
	if err := p.bundleManagerClient.SendDB(ctx, upload.ID, newFilename); err != nil {
		return errors.Wrap(err, "bundleManager.SendDB")
	}

	return nil
}

// updateCommits updates the lsif_commits table with the current data known to gitserver, then updates the
// visibility of all dumps for the given repository.
func (p *processor) updateCommitsAndVisibility(ctx context.Context, db db.DB, repositoryID int, commit string) error {
	tipCommit, err := p.gitserverClient.Head(ctx, db, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}
	newCommits, err := p.gitserverClient.CommitsNear(ctx, db, repositoryID, tipCommit)
	if err != nil {
		return errors.Wrap(err, "gitserver.CommitsNear")
	}

	if tipCommit != commit {
		// If the tip is ahead of this commit, we also want to discover all of the commits between this
		// commit and the tip so that we can accurately determine what is visible from the tip. If we
		// do not do this before the updateDumpsVisibleFromTip call below, no dumps will be reachable
		// from the tip and all dumps will be invisible.
		additionalCommits, err := p.gitserverClient.CommitsNear(ctx, db, repositoryID, commit)
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
	r io.Reader,
	newFilename string,
	dumpID int,
	root string,
	getChildren existence.GetChildrenFunc,
) (_ []types.Package, _ []types.PackageReference, err error) {
	groupedBundleData, err := correlation.Correlate(r, dumpID, root, getChildren)
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
	writer, err := sqlitewriter.NewWriter(filename)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}
	}()

	writers := []func() error{
		func() error {
			return errors.Wrap(writer.WriteMeta(ctx, groupedBundleData.LSIFVersion, groupedBundleData.NumResultChunks), "writer.WriteMeta")
		},
		func() error {
			return errors.Wrap(writer.WriteDocuments(ctx, groupedBundleData.Documents), "writer.WriteDocuments")
		},
		func() error {
			return errors.Wrap(writer.WriteResultChunks(ctx, groupedBundleData.ResultChunks), "writer.WriteResultChunks")
		},
		func() error {
			return errors.Wrap(writer.WriteDefinitions(ctx, groupedBundleData.Definitions), "writer.WriteDefinitions")
		},
		func() error {
			return errors.Wrap(writer.WriteReferences(ctx, groupedBundleData.References), "writer.WriteReferences")
		},
	}

	errs := make(chan error, len(writers))
	defer close(errs)

	for _, w := range writers {
		go func(w func() error) { errs <- w() }(w)
	}

	var writeErr error
	for i := 0; i < len(writers); i++ {
		if err := <-errs; err != nil {
			writeErr = multierror.Append(writeErr, err)
		}
	}
	if writeErr != nil {
		return writeErr
	}

	if err := writer.Flush(ctx); err != nil {
		return errors.Wrap(err, "writer.Flush")
	}

	return nil
}
