package localstore

import (
	"context"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// comments provides access to the `comments` table.
//
// For a detailed overview of the schema, see schema.txt.
type comments struct{}

func (*comments) Create(ctx context.Context, threadID int32, contents string, authorName, authorEmail string) (*sourcegraph.Comment, error) {
	if Mocks.Comments.Create != nil {
		return Mocks.Comments.Create(ctx, threadID, contents, authorName, authorEmail)
	}

	createdAt := time.Now()
	updatedAt := createdAt
	var id int64
	err := appDBH(ctx).QueryRow(
		"INSERT INTO comments(thread_id, contents, created_at, updated_at, author_name, author_email) VALUES($1, $2, $3, $4, $5, $6) RETURNING id",
		threadID, contents, createdAt, updatedAt, authorName, authorEmail).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.Comment{
		ID:          int32(id),
		ThreadID:    threadID,
		Contents:    contents,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		AuthorName:  authorName,
		AuthorEmail: authorEmail,
	}, nil
}

func (c *comments) GetAllForThread(ctx context.Context, threadID int64) ([]*sourcegraph.Comment, error) {
	return c.getBySQL(ctx, "WHERE (thread_id=$1 AND deleted_at IS NULL)", threadID)
}

// getBySQL returns comments matching the SQL query, if any exist.
func (*comments) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Comment, error) {
	rows, err := appDBH(ctx).Query("SELECT id, thread_id, contents, created_at, updated_at, author_name, author_email FROM comments "+query, args...)
	if err != nil {
		return nil, err
	}

	comments := []*sourcegraph.Comment{}
	defer rows.Close()
	for rows.Next() {
		var c sourcegraph.Comment
		err := rows.Scan(&c.ID, &c.ThreadID, &c.Contents, &c.CreatedAt, &c.UpdatedAt, &c.AuthorName, &c.AuthorEmail)
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
