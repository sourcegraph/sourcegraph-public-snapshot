package threads

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type externalThread struct {
	thread        *DBThread
	threadComment commentobjectdb.DBObjectCommentFields

	comments []*comments.ExternalComment
}

var MockImportExternalThreads func(repo api.RepoID, externalServiceID int64, toImport []*externalThread) error

// ImportExternalThreads replaces all existing threads for the objects from the given external
// service with a new set of threads.
func ImportExternalThreads(ctx context.Context, repo api.RepoID, externalServiceID int64, toImport []*externalThread) error {
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
	opt := dbThreadsListOptions{
		RepositoryIDs:                 []api.RepoID{repo},
		ImportedFromExternalServiceID: externalServiceID,
	}

	// Delete all existing threads for the repository from the given external service.
	//
	// TODO!(sqs): if a user replied to an external thread on Sourcegraph and that reply wasnt
	// persisted to the code host, their reply will be cascade-deleted
	if err := (dbThreads{}).Delete(ctx, tx, opt); err != nil {
		return err
	}

	// Insert the new threads.
	for _, x := range toImport {
		if x.thread.ImportedFromExternalServiceID != externalServiceID {
			panic("external service ID mismatch")
		}
		if x.thread.ExternalID == "" {
			panic("thread has no external ID")
		}
		if _, err := dbCreateExternalThread(ctx, tx, x); err != nil {
			return err
		}
	}
	return nil
}

func dbCreateExternalThread(ctx context.Context, tx *sql.Tx, x *externalThread) (threadID int64, err error) {
	dbThread, err := (dbThreads{}).Create(ctx, tx, x.thread, x.threadComment)
	if err != nil {
		return 0, err
	}
	for _, comment := range x.comments {
		tmp := *comment
		tmp.ThreadPrimaryCommentID = dbThread.PrimaryCommentID
		if err := comments.CreateExternalCommentReply(ctx, tx, tmp); err != nil {
			return 0, err
		}
	}
	return dbThread.ID, nil
}

func dbUpdateExternalThread(ctx context.Context, threadID int64, x *externalThread) error {
	update := dbThreadUpdate{
		Title: &x.thread.Title,
		State: &x.thread.State,
	}
	if x.thread.BaseRef != "" {
		update.BaseRef = &x.thread.BaseRef
	}
	if x.thread.HeadRef != "" {
		update.HeadRef = &x.thread.HeadRef
	}
	dbThread, err := (dbThreads{}).Update(ctx, threadID, update)
	if err != nil {
		return err
	}

	// TODO!(sqs): hack: to avoid duplicating comments, we delete all comments for the thread and then
	// re-add all. this will result in permanent removal of non-external comments.
	if _, err := dbconn.Global.ExecContext(ctx, `DELETE FROM comments WHERE parent_comment_id=$1`, dbThread.PrimaryCommentID); err != nil {
		return err
	}

	for _, comment := range x.comments {
		tmp := *comment
		tmp.ThreadPrimaryCommentID = dbThread.PrimaryCommentID
		if err := comments.CreateExternalCommentReply(ctx, nil, tmp); err != nil {
			return err
		}
	}
	return nil
}
