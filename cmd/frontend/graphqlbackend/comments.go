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

func (schemaResolver) Comments(ctx context.Context, arg *graphqlutil.ConnectionArgs) (CommentConnection, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.Comments(ctx, arg)
}

func (r schemaResolver) CreateComment(ctx context.Context, arg *CreateCommentArgs) (Comment, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.CreateComment(ctx, arg)
}

func (r schemaResolver) EditComment(ctx context.Context, arg *EditCommentArgs) (Comment, error) {
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

// CommentsForObject returns an instance of the GraphQL CommentConnection type with the list of
// comments on an object.
func CommentsForObject(ctx context.Context, object graphql.ID, arg *graphqlutil.ConnectionArgs) (CommentConnection, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	return Comments.CommentsForObject(ctx, object, arg)
}

// CommentsResolver is the interface for the GraphQL comments queries and mutations.
type CommentsResolver interface {
	// Queries
	Comments(context.Context, *graphqlutil.ConnectionArgs) (CommentConnection, error)

	// Mutations
	CreateComment(context.Context, *CreateCommentArgs) (Comment, error)
	EditComment(context.Context, *EditCommentArgs) (Comment, error)
	DeleteComment(context.Context, *DeleteCommentArgs) (*EmptyResponse, error)

	// CommentsForObject is called by the CommentsForObject func but is not in the GraphQL
	// API.
	CommentsForObject(ctx context.Context, object graphql.ID, arg *graphqlutil.ConnectionArgs) (CommentConnection, error)
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

// PartialComment is the interface for the GraphQL type Comment.
//
// The ID field is required in the GraphQL interface but omitted in the Go interface to avoid Go
// compiler errors that are not helpful.
type PartialComment interface {
	Author(context.Context) (*Actor, error)
	Body(context.Context) (string, error)
	CreatedAt(context.Context) (DateTime, error)
	UpdatedAt(context.Context) (DateTime, error)
}

type Comment interface {
	ID() graphql.ID
	PartialComment
	ToThread() (Thread, bool)
	ToChangeset() (Changeset, bool)
	ToCampaign() (Campaign, bool)
}

type ToComment struct {
	Thread    Thread
	Changeset Changeset
	Campaign  Campaign
}

func (v ToComment) comment() interface {
	ID() graphql.ID
	PartialComment
} {
	switch {
	case v.Thread != nil:
		return v.Thread
		// TODO!(sqs): add Issue
	case v.Changeset != nil:
		return v.Changeset
	case v.Campaign != nil:
		return v.Campaign
	default:
		panic("invalid ToComment")
	}
}

func (v ToComment) ID() graphql.ID                                  { return v.comment().ID() }
func (v ToComment) Author(ctx context.Context) (*Actor, error)      { return v.comment().Author(ctx) }
func (v ToComment) Body(ctx context.Context) (string, error)        { return v.comment().Body(ctx) }
func (v ToComment) UpdatedAt(ctx context.Context) (DateTime, error) { return v.comment().UpdatedAt(ctx) }
func (v ToComment) CreatedAt(ctx context.Context) (DateTime, error) { return v.comment().CreatedAt(ctx) }
func (v ToComment) ToThread() (Thread, bool)                        { return v.Thread, v.Thread != nil }
func (v ToComment) ToChangeset() (Changeset, bool)                  { return v.Changeset, v.Changeset != nil }
func (v ToComment) ToCampaign() (Campaign, bool)                    { return v.Campaign, v.Campaign != nil }

// CommentConnection is the interface for the GraphQL type CommentConnection.
type CommentConnection interface {
	Nodes(context.Context) ([]Comment, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
