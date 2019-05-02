package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// TODO(slimsag:discussions): future: tests for DiscussionComments.List
// TODO(slimsag:discussions): future: tests for DiscussionComments.Count

// discussionComments provides access to the `discussion_comments` table.
//
// For a detailed overview of the schema, see schema.md.
type discussionComments struct{}

// ErrCommentNotFound is the error returned by Discussions methods to indicate
// that the comment could not be found.
type ErrCommentNotFound struct {
	// CommentID is the comment that was not found.
	CommentID int64
}

func (e *ErrCommentNotFound) Error() string {
	return fmt.Sprintf("comment %d not found", e.CommentID)
}

func (c *discussionComments) Create(ctx context.Context, newComment *types.DiscussionComment) (*types.DiscussionComment, error) {
	if Mocks.DiscussionComments.Create != nil {
		return Mocks.DiscussionComments.Create(ctx, newComment)
	}

	// Validate the input comment.
	if newComment == nil {
		return nil, errors.New("newComment is nil")
	}
	if newComment.ID != 0 {
		return nil, errors.New("newComment.ID must be zero")
	}
	if len([]rune(newComment.Contents)) > 100000 {
		return nil, errors.New("comment content too long (must be less than 100,000 UTF-8 characters)")
	}
	if !newComment.CreatedAt.IsZero() {
		return nil, errors.New("newComment.CreatedAt must not be specified")
	}
	if !newComment.UpdatedAt.IsZero() {
		return nil, errors.New("newComment.UpdatedAt must not be specified")
	}
	if newComment.DeletedAt != nil {
		return nil, errors.New("newComment.DeletedAt must not be specified")
	}

	// Create the comment.
	newComment.CreatedAt = time.Now()
	newComment.UpdatedAt = newComment.CreatedAt

	err := dbconn.Global.QueryRowContext(ctx, `INSERT INTO discussion_comments(
		thread_id,
		author_user_id,
		contents,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		newComment.ThreadID,
		newComment.AuthorUserID,
		newComment.Contents,
		newComment.CreatedAt,
		newComment.UpdatedAt,
	).Scan(&newComment.ID)
	if err != nil {
		return nil, err
	}
	return newComment, nil
}

type DiscussionCommentsUpdateOptions struct {
	// Contents, when non-nil, specifies the new contents of the comment.
	Contents *string

	// Delete, when true, specifies that the comment should be deleted. This
	// operation cannot be undone.
	Delete bool

	// Report, when non-nil, specifies that the report message string should be
	// added to the list of reports on this comment.
	Report *string

	// ClearReports, when true, specifies that the comments reports should be
	// cleared (e.g. after review by an admin)
	ClearReports bool

	// noThreadDelete prevents calling DiscussionThreads.Delete when the comment
	// being deleted is the first comment in the thread. This should ONLY be
	// used by DiscussionThreads.Delete to avoid circular calls.
	noThreadDelete bool
}

func (c *discussionComments) Update(ctx context.Context, commentID int64, opts *DiscussionCommentsUpdateOptions) (*types.DiscussionComment, error) {
	if Mocks.DiscussionComments.Update != nil {
		return Mocks.DiscussionComments.Update(ctx, commentID, opts)
	}
	if opts == nil {
		return nil, errors.New("options must not be nil")
	}
	now := time.Now()

	// TODO(slimsag:discussions): should be in a transaction

	anyUpdate := false
	if opts.Contents != nil {
		anyUpdate = true
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_comments SET contents=$1 WHERE id=$2 AND deleted_at IS NULL", *opts.Contents, commentID); err != nil {
			return nil, err
		}
	}
	var deletingFirstComment bool
	if opts.Delete {
		// Deleting the first comment in a thread implicitly means deleting the thread itself.
		var (
			deletingFirstComment bool
			threadID             int64
		)
		if !opts.noThreadDelete {
			comment, err := c.Get(ctx, commentID)
			if err != nil {
				return nil, err
			}
			comments, err := c.List(ctx, &DiscussionCommentsListOptions{
				ThreadID: &comment.ThreadID,
			})
			if err != nil {
				return nil, err
			}
			deletingFirstComment = comment.ID == comments[0].ID
			threadID = comment.ThreadID
		}
		if deletingFirstComment {
			_, err := DiscussionThreads.Update(ctx, threadID, &DiscussionThreadsUpdateOptions{Delete: true})
			if err != nil {
				return nil, err
			}
		}
	}
	if !deletingFirstComment && opts.Delete {
		anyUpdate = true
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_comments SET deleted_at=$1 WHERE id=$2 AND deleted_at IS NULL", now, commentID); err != nil {
			return nil, err
		}
	}
	if opts.Report != nil {
		anyUpdate = true
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_comments SET reports=ARRAY_APPEND(reports,$1) WHERE id=$2 AND deleted_at IS NULL", *opts.Report, commentID); err != nil {
			return nil, err
		}
	}
	if opts.ClearReports {
		anyUpdate = true
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_comments SET reports='{}' WHERE id=$1 AND deleted_at IS NULL", commentID); err != nil {
			return nil, err
		}
	}
	if anyUpdate {
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_comments SET updated_at=$1 WHERE id=$2 AND deleted_at IS NULL", now, commentID); err != nil {
			return nil, err
		}
	}
	if opts.Delete {
		return nil, nil
	}
	return c.Get(ctx, commentID)
}

type DiscussionCommentsListOptions struct {
	// LimitOffset specifies SQL LIMIT and OFFSET counts. It may be nil (no limit / offset).
	*LimitOffset

	// AuthorUserID, when non-nil, specifies that only comments made by this
	// author should be returned.
	AuthorUserID *int32

	// ThreadID, when non-nil, specifies that only comments in this thread ID
	// should be returned.
	ThreadID *int64

	// CommentID, when non-nil, specifies that only comments with this ID should
	// be returned.
	CommentID *int64

	// Reported, when true, returns only threads that have at least one report.
	Reported bool

	// CreatedBefore, when non-nil, specifies that only comments that were
	// created before this time should be returned.
	CreatedBefore *time.Time
	CreatedAfter  *time.Time
}

func (c *discussionComments) List(ctx context.Context, opts *DiscussionCommentsListOptions) ([]*types.DiscussionComment, error) {
	if Mocks.DiscussionComments.List != nil {
		return Mocks.DiscussionComments.List(ctx, opts)
	}
	if opts == nil {
		return nil, errors.New("options must not be nil")
	}
	conds := c.getListSQL(opts)
	q := sqlf.Sprintf("WHERE %s ORDER BY id ASC %s", sqlf.Join(conds, "AND"), opts.LimitOffset.SQL())
	return c.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (c *discussionComments) Get(ctx context.Context, commentID int64) (*types.DiscussionComment, error) {
	comments, err := c.List(ctx, &DiscussionCommentsListOptions{
		CommentID: &commentID,
	})
	if err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return nil, &ErrCommentNotFound{CommentID: commentID}
	}
	return comments[0], nil
}

func (c *discussionComments) Count(ctx context.Context, opts *DiscussionCommentsListOptions) (int, error) {
	if Mocks.DiscussionComments.Count != nil {
		return Mocks.DiscussionComments.Count(ctx, opts)
	}
	if opts == nil {
		return 0, errors.New("options must not be nil")
	}
	conds := c.getListSQL(opts)
	q := sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "AND"))
	return c.getCountBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (*discussionComments) getListSQL(opts *DiscussionCommentsListOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	conds = append(conds, sqlf.Sprintf("deleted_at IS NULL"))
	if opts.AuthorUserID != nil {
		conds = append(conds, sqlf.Sprintf("author_user_id=%v", *opts.AuthorUserID))
	}
	if opts.ThreadID != nil {
		conds = append(conds, sqlf.Sprintf("thread_id=%v", *opts.ThreadID))
	}
	if opts.CommentID != nil {
		conds = append(conds, sqlf.Sprintf("id=%v", *opts.CommentID))
	}
	if opts.Reported {
		conds = append(conds, sqlf.Sprintf("array_length(reports,1) > 0"))
	}
	if opts.CreatedBefore != nil {
		conds = append(conds, sqlf.Sprintf("created_at < %v", *opts.CreatedBefore))
	}
	if opts.CreatedAfter != nil {
		conds = append(conds, sqlf.Sprintf("created_at > %v", *opts.CreatedAfter))
	}
	return conds
}

func (*discussionComments) getCountBySQL(ctx context.Context, query string, args ...interface{}) (int, error) {
	var count int
	rows := dbconn.Global.QueryRowContext(ctx, "SELECT count(id) FROM discussion_comments t "+query, args...)
	err := rows.Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// getBySQL returns comments matching the SQL query, if any exist.
func (*discussionComments) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.DiscussionComment, error) {
	rows, err := dbconn.Global.QueryContext(ctx, `
		SELECT
			c.id,
			c.thread_id,
			c.author_user_id,
			c.contents,
			c.created_at,
			c.updated_at,
			c.reports
		FROM discussion_comments c `+query, args...)
	if err != nil {
		return nil, err
	}

	comments := []*types.DiscussionComment{}
	defer rows.Close()
	for rows.Next() {
		comment := &types.DiscussionComment{}
		err := rows.Scan(
			&comment.ID,
			&comment.ThreadID,
			&comment.AuthorUserID,
			&comment.Contents,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			pq.Array(&comment.Reports),
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}
