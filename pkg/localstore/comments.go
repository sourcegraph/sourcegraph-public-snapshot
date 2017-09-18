package localstore

import (
	"context"
	"database/sql"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

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

	createdAt := time.Now()
	updatedAt := createdAt
	var id int32
	var err error
	if authorUserID != "" {
		err = globalDB.QueryRow(
			"INSERT INTO comments(thread_id, contents, created_at, updated_at, author_user_id) VALUES($1, $2, $3, $4, $5) RETURNING id",
			threadID, contents, createdAt, updatedAt, authorUserID).Scan(&id)
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

func (c *comments) GetAllForThread(ctx context.Context, threadID int32) ([]*sourcegraph.Comment, error) {
	return c.getBySQL(ctx, "WHERE (thread_id=$1 AND deleted_at IS NULL)", threadID)
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
		} else if authorName.Valid && authorEmail.Valid {
			// deprecated code path
			c.AuthorName = authorName.String
			c.AuthorEmail = authorEmail.String
		} else {
			// Skip invalid db row
			log15.Warn("invalid comment in database", "id", c.ID)
			continue
		}
		comments = append(comments, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
