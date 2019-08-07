package threads

import (
	"context"

	"github.com/hashicorp/go-multierror"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

var MockImportExternalThreads func(repo api.RepoID, externalServiceID int64, toImport map[*DBThread]commentobjectdb.DBObjectCommentFields) error

// ImportExternalThreads replaces all existing threads for the objects from the given external service
// with a new set of threads.
func ImportExternalThreads(ctx context.Context, repo api.RepoID, externalServiceID int64, toImport map[*DBThread]commentobjectdb.DBObjectCommentFields) error {
	if MockImportExternalThreads != nil {
		return MockImportExternalThreads(repo, externalServiceID, toImport)
	}

	tx, err := dbconn.Global.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	if externalServiceID == 0 {
		panic("externalServiceID must be nonzero")
	}
	opt := DBThreadsListOptions{
		RepositoryID:                  repo,
		ImportedFromExternalServiceID: externalServiceID,
	}

	// Delete all existing threads for the repository from the given external service.
	if err := (DBThreads{}).Delete(ctx, tx, opt); err != nil {
		return err
	}

	// Insert the new threads.
	for thread, comment := range toImport {
		if thread.ImportedFromExternalServiceID != externalServiceID {
			panic("external service ID mismatch")
		}
		if thread.ExternalID == "" {
			panic("thread has no external ID")
		}
		if _, err := (DBThreads{}).Create(ctx, tx, thread, comment); err != nil {
			return err
		}
	}
	return nil
}
