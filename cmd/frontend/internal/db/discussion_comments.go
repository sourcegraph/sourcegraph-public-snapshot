package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/pkg/types"
)

// TODO(slimsag:discussions): future: tests for DiscussionComments.List
// TODO(slimsag:discussions): future: tests for DiscussionComments.Count

// discussionComments provides access to the `discussion_comments` table.
//
// For a detailed overview of the schema, see schema.md.
type discussionComments struct{}

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

	err := globalDB.QueryRowContext(ctx, `INSERT INTO discussion_comments(
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

type DiscussionCommentsListOptions struct {
	// LimitOffset specifies SQL LIMIT and OFFSET counts. It may be nil (no limit / offset).
	*LimitOffset

	// AuthorUserID, when non-nil, specifies that only comments made by this
	// author should be returned.
	AuthorUserID *int32

	// ThreadID, when non-nil, specifies that only comments in this thread ID
	// should be returned.
	ThreadID *int64
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
	return conds
}

func (*discussionComments) getCountBySQL(ctx context.Context, query string, args ...interface{}) (int, error) {
	var count int
	rows := globalDB.QueryRowContext(ctx, "SELECT count(id) FROM discussion_comments t "+query, args...)
	err := rows.Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// getBySQL returns comments matching the SQL query, if any exist.
func (*discussionComments) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.DiscussionComment, error) {
	rows, err := globalDB.QueryContext(ctx, `
		SELECT
			c.id,
			c.thread_id,
			c.author_user_id,
			c.contents,
			c.created_at,
			c.updated_at
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
