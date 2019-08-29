package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
)

// Comments is the implementation of the GraphQL comments queries and mutations. If it is not set at
// runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Comments CommentsResolver

var errCommentsNotImplemented = errors.New("comments is not implemented")

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

// CommentsResolver is the interface for the GraphQL comments queries and mutations.
type CommentsResolver interface {
	// Mutations
	EditComment(context.Context, *EditCommentArgs) (Comment, error)
	DeleteComment(context.Context, *DeleteCommentArgs) (*EmptyResponse, error)
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
	BodyText(context.Context) (string, error)
	BodyHTML(context.Context) (string, error)
	CreatedAt(context.Context) (DateTime, error)
	UpdatedAt(context.Context) (DateTime, error)
}

type Comment interface {
	ID() graphql.ID
	PartialComment
	ToCampaign() (Campaign, bool)
}

type ToComment struct {
	Campaign Campaign

	// In the future, CommentReply and Thread types will be supported.
}

func (v ToComment) comment() interface {
	ID() graphql.ID
	PartialComment
} {
	switch {
	case v.Campaign != nil:
		return v.Campaign
	default:
		panic("invalid Comment")
	}
}

func (v ToComment) ID() graphql.ID                                  { return v.comment().ID() }
func (v ToComment) Author(ctx context.Context) (*Actor, error)      { return v.comment().Author(ctx) }
func (v ToComment) Body(ctx context.Context) (string, error)        { return v.comment().Body(ctx) }
func (v ToComment) BodyText(ctx context.Context) (string, error)    { return v.comment().BodyText(ctx) }
func (v ToComment) BodyHTML(ctx context.Context) (string, error)    { return v.comment().BodyHTML(ctx) }
func (v ToComment) UpdatedAt(ctx context.Context) (DateTime, error) { return v.comment().UpdatedAt(ctx) }
func (v ToComment) CreatedAt(ctx context.Context) (DateTime, error) { return v.comment().CreatedAt(ctx) }
func (v ToComment) ToCampaign() (Campaign, bool)                    { return v.Campaign, v.Campaign != nil }

type CommentEvent struct {
	EventCommon
	// TODO!(sqs): include the comment
}
