package localstore

import (
	"context"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func init() {
	AppSchema.Map.AddTableWithName(dbComment{}, "comments").SetKeys(true, "ID")
}

// dbComment DB-maps a sourcegraph.Comment object.
type dbComment struct {
	ID          int64
	ThreadID    int64 `db:"thread_id"`
	Contents    string
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	AuthorName  string     `db:"author_name"`
	AuthorEmail string     `db:"author_email"`
}

type comments struct{}

func (c *dbComment) fromComment(c2 *sourcegraph.Comment) {
	c.ID = int64(c2.ID)
	c.ThreadID = int64(c2.ThreadID)
	c.Contents = c2.Contents
	c.AuthorName = c2.AuthorName
	c.AuthorEmail = c2.AuthorEmail
}

func (c *dbComment) toComment() *sourcegraph.Comment {
	c2 := &sourcegraph.Comment{}
	c2.ID = int32(c.ID)
	c2.ThreadID = int32(c.ThreadID)
	c2.Contents = c.Contents
	c2.CreatedAt = c.CreatedAt
	c2.UpdatedAt = c.UpdatedAt
	c2.AuthorName = c.AuthorName
	c2.AuthorEmail = c.AuthorEmail
	return c2
}

func (*comments) Create(ctx context.Context, newComment *sourcegraph.Comment) (*sourcegraph.Comment, error) {
	if Mocks.Comments.Create != nil {
		return Mocks.Comments.Create(ctx, newComment)
	}

	var c dbComment
	c.fromComment(newComment)
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	err := appDBH(ctx).Insert(&c)
	if err != nil {
		return nil, err
	}

	return c.toComment(), nil
}

func (*comments) GetAllForThread(ctx context.Context, threadID int64) ([]*sourcegraph.Comment, error) {
	cs := []*dbComment{}
	_, err := appDBH(ctx).Select(&cs, "SELECT * FROM comments WHERE (thread_id=$1 AND deleted_at IS NULL)", threadID)
	if err != nil {
		return nil, err
	}

	comments := []*sourcegraph.Comment{}
	for _, c := range cs {
		comments = append(comments, c.toComment())
	}
	return comments, nil
}
