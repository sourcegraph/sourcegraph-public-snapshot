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

// CommentsForObject returns an instance of the GraphQL CommentConnection type with the list of
// comments on an object.
func CommentsForObject(ctx context.Context, object graphql.ID, arg *graphqlutil.ConnectionArgs) (CommentConnection, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.CommentsForObject(ctx, object, arg)
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
//
// The ID field is required in the GraphQL interface but omitted in the Go interface to avoid Go
// compiler errors that are not helpful.
type Comment interface {
	Author(context.Context) (*Actor, error)
	Body() string
	CreatedAt() DateTime
	UpdatedAt() DateTime
}

type ToComment struct {
	Thread    Thread
	Changeset Changeset
}

var (
	_ Comment = Thread(nil)
	_ Comment = Changeset(nil)
)

func (v ToComment) comment() interface {
	ID() graphql.ID
	Comment
} {
	switch {
	case v.Thread != nil:
		return v.Thread
		// TODO!(sqs): add Issue
	case v.Changeset != nil:
		return v.Changeset
	default:
		panic("invalid ToComment")
	}
}

func (v ToComment) ID() graphql.ID                             { return v.comment().ID() }
func (v ToComment) Author(ctx context.Context) (*Actor, error) { return v.comment().Author(ctx) }
func (v ToComment) Body() string                               { return v.comment().Body() }
func (v ToComment) CreatedAt() DateTime                        { return v.comment().CreatedAt() }
func (v ToComment) UpdatedAt() DateTime                        { return v.comment().UpdatedAt() }

func (v ToComment) ToThread() (Thread, bool)       { return v.Thread, v.Thread != nil }
func (v ToComment) ToChangeset() (Changeset, bool) { return v.Changeset, v.Changeset != nil }

// CommentConnection is the interface for the GraphQL type CommentConnection.
type CommentConnection interface {
	Nodes(context.Context) ([]ToComment, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
