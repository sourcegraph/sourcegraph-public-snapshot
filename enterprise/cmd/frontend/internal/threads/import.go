package threads

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type externalThread struct {
	thread        *DBThread
	threadComment commentobjectdb.DBObjectCommentFields

	comments []comments.ExternalComment
}

func dbCreateExternalThread(ctx context.Context, tx *sql.Tx, x externalThread) (threadID int64, err error) {
	dbThread, err := (dbThreads{}).Create(ctx, tx, x.thread, x.threadComment)
	if err != nil {
		return 0, err
	}
	for _, comment := range x.comments {
		comment.ThreadPrimaryCommentID = dbThread.PrimaryCommentID
		if err := comments.CreateExternalCommentReply(ctx, tx, comment); err != nil {
			return 0, err
		}
	}
	return dbThread.ID, nil
}

func dbUpdateExternalThread(ctx context.Context, threadID int64, x externalThread) error {
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
		comment.ThreadPrimaryCommentID = dbThread.PrimaryCommentID
		if err := comments.CreateExternalCommentReply(ctx, nil, comment); err != nil {
			return err
		}
	}
	return nil
}
