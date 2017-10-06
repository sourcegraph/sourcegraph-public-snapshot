package localstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// comments provides access to the `comments` table.
//
// For a detailed overview of the schema, see schema.txt.
type comments struct{}

func (*comments) Create(ctx context.Context, threadID int32, contents, authorName, authorEmail, authorUserID string) (*sourcegraph.Comment, error) {
	if Mocks.Comments.Create != nil {
		return Mocks.Comments.Create(ctx, threadID, contents, authorName, authorEmail)
	}

	if len(contents) > 100000 {
		return nil, errors.New("comment too long")
	}

	createdAt := time.Now()
	updatedAt := createdAt
	var id int32
	var err error
	if authorUserID != "" {
		err = globalDB.QueryRow(
			"INSERT INTO comments(thread_id, contents, created_at, updated_at, author_user_id, author_name, author_email) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id",
			threadID, contents, createdAt, updatedAt, authorUserID, authorName, authorEmail).Scan(&id)
	} else {
		// deprecated code path
		err = globalDB.QueryRow(
			"INSERT INTO comments(thread_id, contents, created_at, updated_at, author_name, author_email) VALUES($1, $2, $3, $4, $5, $6) RETURNING id",
			threadID, contents, createdAt, updatedAt, authorName, authorEmail).Scan(&id)
	}
	if err != nil {
		return nil, err
	}

	return &sourcegraph.Comment{
		ID:           id,
		ThreadID:     threadID,
		Contents:     contents,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		AuthorName:   authorName,
		AuthorEmail:  authorEmail,
		AuthorUserID: authorUserID,
	}, nil
}

func (c *comments) GetByID(ctx context.Context, commentID int32) (*sourcegraph.Comment, error) {
	comments, err := c.getBySQL(ctx, "WHERE id=$1 AND deleted_at IS NULL ORDER BY id ASC", commentID)
	if err != nil {
		return nil, err
	}
	if len(comments) != 1 {
		return nil, fmt.Errorf("comment ID %d does not exist", commentID)
	}
	return comments[0], nil
}

func (c *comments) GetAllForThread(ctx context.Context, threadID int32) ([]*sourcegraph.Comment, error) {
	return c.getBySQL(ctx, "WHERE thread_id=$1 AND deleted_at IS NULL ORDER BY id ASC", threadID)
}

// getBySQL returns comments matching the SQL query, if any exist.
func (*comments) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Comment, error) {
	rows, err := globalDB.Query("SELECT id, thread_id, author_user_id, contents, created_at, updated_at, author_name, author_email FROM comments "+query, args...)
	if err != nil {
		return nil, err
	}

	comments := []*sourcegraph.Comment{}
	defer rows.Close()
	for rows.Next() {
		var c sourcegraph.Comment
		var authorUserID, authorName, authorEmail sql.NullString
		err := rows.Scan(&c.ID, &c.ThreadID, &authorUserID, &c.Contents, &c.CreatedAt, &c.UpdatedAt, &authorName, &authorEmail)
		if err != nil {
			return nil, err
		}
		if authorUserID.Valid {
			c.AuthorUserID = authorUserID.String
		}
		if authorName.Valid {
			c.AuthorName = authorName.String
		}
		if authorEmail.Valid {
			c.AuthorEmail = authorEmail.String
		}
		comments = append(comments, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
