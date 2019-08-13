package commentobjectdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/internal"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type insertRelatedObjectFunc func(ctx context.Context, tx *sql.Tx, commentID int64) (*types.CommentObject, error)

// DBObjectCommentFields contains the subset of fields required when creating a comment that is
// related to another object.
type DBObjectCommentFields struct {
	Author    actor.DBColumns
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateCommentWithObject creates a comment and its related object (such as a thread or campaign)
// in a transaction.
//
// The insertRelatedObject func is called with the comment ID, in case the related object needs to
// store the comment ID. After the related object is inserted, the comment row is updated to refer
// to the object by ID (e.g., the thread ID or campaign ID).
func CreateCommentWithObject(ctx context.Context, tx *sql.Tx, comment DBObjectCommentFields, insertRelatedObject insertRelatedObjectFunc) (err error) {
	if tx == nil {
		var err error
		tx, err = dbconn.Global.BeginTx(ctx, nil)
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
	}

	dbComment := &internal.DBComment{
		Author:    comment.Author,
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
	insertedComment, err := internal.DBComments{}.Create(ctx, tx, dbComment)
	if err != nil {
		return err
	}

	object, err := insertRelatedObject(ctx, tx, insertedComment.ID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE comments SET parent_comment_id=$1, thread_id=$2, campaign_id=$3 WHERE id=$4`,
		nilIfZero(object.ParentCommentID),
		nilIfZero(object.ThreadID),
		nilIfZero(object.CampaignID),
		insertedComment.ID,
	)
	return err
}

func nilIfZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
