package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Comments is the implementation of the GraphQL comments queries and mutations. If it is not set at
// runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Comments CommentsResolver

var errCommentsNotImplemented = errors.New("comments is not implemented")

// CommentByID is called to look up a Comment given its GraphQL ID.
func CommentByID(ctx context.Context, id graphql.ID) (*ToComment, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.CommentByID(ctx, id)
}

// CommentsForObject returns an instance of the GraphQL CommentConnection type with the list of
// comments defined in a namespace.
func CommentsForObject(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (CommentConnection, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.CommentsForObject(ctx, namespace, arg)
}

func (schemaResolver) Comments(ctx context.Context, arg *graphqlutil.ConnectionArgs) (CommentConnection, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.Comments(ctx, arg)
}

func (r schemaResolver) CreateComment(ctx context.Context, arg *CreateCommentArgs) (*ToComment, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.CreateComment(ctx, arg)
}

func (r schemaResolver) EditComment(ctx context.Context, arg *EditCommentArgs) (*ToComment, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.EditComment(ctx, arg)
}

func (r schemaResolver) DeleteComment(ctx context.Context, arg *DeleteCommentArgs) (*EmptyResponse, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.DeleteComment(ctx, arg)
}

// CommentsResolver is the interface for the GraphQL comments queries and mutations.
type CommentsResolver interface {
	// Queries
	Comments(context.Context, *graphqlutil.ConnectionArgs) (CommentConnection, error)

	// Mutations
	CreateComment(context.Context, *CreateCommentArgs) (*ToComment, error)
	EditComment(context.Context, *EditCommentArgs) (*ToComment, error)
	DeleteComment(context.Context, *DeleteCommentArgs) (*EmptyResponse, error)

	// CommentByID is called by the CommentByID func but is not in the GraphQL API.
	CommentByID(context.Context, graphql.ID) (*ToComment, error)

	// CommentsForObject is called by the CommentsForObject func but is not in the GraphQL
	// API.
	CommentsForObject(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (CommentConnection, error)
}

type CreateCommentArgs struct {
	Input struct {
		Node graphql.ID
		Body string
	}
}

type EditCommentArgs struct {
	Input struct {
		ID   graphql.ID
		Body string
	}
}

type DeleteCommentArgs struct {
	Comment graphql.ID
}

// Comment is the interface for the GraphQL type Comment.
type Comment interface {
	ID() graphql.ID
	Author() *Actor
	Body() string
	CreatedAt() DateTime
	UpdatedAt() DateTime
}

type ToComment struct {
	Thread    Thread
	Changeset Changeset
}

func (v ToComment) ToThread() (Thread, bool)       { return v.Thread, v.Thread != nil }
func (v ToComment) ToChangeset() (Changeset, bool) { return v.Changeset, v.Changeset != nil }

// CommentConnection is the interface for the GraphQL type CommentConnection.
type CommentConnection interface {
	Nodes(context.Context) ([]ToComment, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
