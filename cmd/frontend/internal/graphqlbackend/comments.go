package graphqlbackend

import (
	"context"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type commentResolver struct {
	comment *sourcegraph.Comment
}

func (c *commentResolver) ID() int32 {
	return c.comment.ID
}

func (c *commentResolver) Contents() string {
	return c.comment.Contents
}

func (c *commentResolver) AuthorName() string {
	return c.comment.AuthorName
}

func (c *commentResolver) AuthorEmail() string {
	return c.comment.AuthorEmail
}

func (c *commentResolver) CreatedAt() string {
	return c.comment.CreatedAt.Format(time.RFC3339) // ISO
}

func (c *commentResolver) UpdatedAt() string {
	return c.comment.UpdatedAt.Format(time.RFC3339) // ISO
}

func (*schemaResolver) AddCommentToThread(ctx context.Context, args *struct {
	RemoteURI   string
	AccessToken string
	ThreadID    int32
	Contents    string
	AuthorName  string
	AuthorEmail string
}) (*threadResolver, error) {
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! LocalRepos.Get is responsible for 🚨
	// ensuring the user has permissions to access the repository.
	_, err := store.LocalRepos.Get(ctx, args.RemoteURI, args.AccessToken)
	if err != nil {
		return nil, err
	}

	thread, err := store.Threads.Get(ctx, int64(args.ThreadID))
	if err != nil {
		return nil, err
	}

	_, err = store.Comments.Create(ctx, &sourcegraph.Comment{
		ThreadID:    args.ThreadID,
		Contents:    args.Contents,
		AuthorName:  args.AuthorName,
		AuthorEmail: args.AuthorEmail,
	})
	if err != nil {
		return nil, err
	}

	return &threadResolver{thread: thread}, nil
}
