package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// comments provides access to the `comments` table.
//
// For a detailed overview of the schema, see schema.txt.
type comments struct{}

func (*comments) Create(ctx context.Context, threadID int32, contents string, authorUserID int32) (*sourcegraph.Comment, error) {
	if Mocks.Comments.Create != nil {
		return Mocks.Comments.Create(ctx, threadID, contents, authorUserID)
	}

	if len(contents) > 100000 {
		return nil, errors.New("comment too long")
	}

	createdAt := time.Now()
	updatedAt := createdAt
	var id int32
	if authorUserID == 0 {
		return nil, errors.New("must specify author ID to create comment")
	}
	err := globalDB.QueryRowContext(
		ctx,
		"INSERT INTO comments(thread_id, contents, created_at, updated_at, author_user_id) VALUES($1, $2, $3, $4, $5) RETURNING id",
		threadID, contents, createdAt, updatedAt, authorUserID).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.Comment{
		ID:           id,
		ThreadID:     threadID,
		Contents:     contents,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		AuthorUserID: authorUserID,
	}, nil
}

func (c *comments) GetByID(ctx context.Context, commentID int32) (*sourcegraph.Comment, error) {
	comments, err := c.getBySQL(ctx, "WHERE id=$1 AND deleted_at IS NULL", commentID)
	if err != nil {
		return nil, err
	}
	if len(comments) != 1 {
		return nil, fmt.Errorf("comment ID %d does not exist", commentID)
	}
	return comments[0], nil
}

func (c *comments) GetAllForThread(ctx context.Context, threadID int32) ([]*sourcegraph.Comment, error) {
	if Mocks.Comments.GetAllForThread != nil {
		return Mocks.Comments.GetAllForThread(ctx, threadID)
	}

	return c.getBySQL(ctx, "WHERE thread_id=$1 AND deleted_at IS NULL ORDER BY id ASC", threadID)
}

// getBySQL returns comments matching the SQL query, if any exist.
func (*comments) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Comment, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT id, thread_id, author_user_id, contents, created_at, updated_at FROM comments "+query, args...)
	if err != nil {
		return nil, err
	}

	comments := []*sourcegraph.Comment{}
	defer rows.Close()
	for rows.Next() {
		var c sourcegraph.Comment
		err := rows.Scan(&c.ID, &c.ThreadID, &c.AuthorUserID, &c.Contents, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
