package graphqlbackend

import (
	"context"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (schemaResolver) Commentable(ctx context.Context, arg *struct{ ID graphql.ID }) (*ToCommentable, error) {
	if Comments == nil {
		return nil, errCommentsNotImplemented
	}
	node, err := NodeByID(ctx, arg.ID)
	if err != nil {
		return nil, err
	}

	var toCommentable ToCommentable
	switch relay.UnmarshalKind(arg.ID) {
	// TODO!(sqs): support nested comments? comments on comments?
	case GQLTypeCampaign:
		toCommentable.Campaign = node.(Campaign)
	case GQLTypeThread:
		toCommentable.Thread = node.(Thread)
	default:
		return nil, fmt.Errorf("node %q is not commentable", arg.ID)
	}
	return &toCommentable, nil
}

type CannotCommentReason string

const (
	CannotCommentReasonAuthenticationRequired = "AUTHENTICATION_REQUIRED"
)

// commentable is the interface for the GraphQL interface Commentable.
type commentable interface {
	ViewerCanComment(context.Context) (bool, error)
	ViewerCannotCommentReasons(context.Context) ([]CannotCommentReason, error)
	Comments(context.Context, *graphqlutil.ConnectionArgs) (CommentConnection, error)
}

type ToCommentable struct {
	Campaign Campaign
	Thread   Thread
}

func (v ToCommentable) commentable() commentable {
	switch {
	case v.Campaign != nil:
		return v.Campaign
	case v.Thread != nil:
		return v.Thread
	default:
		panic("invalid Commentable")
	}
}

func (v ToCommentable) ViewerCanComment(ctx context.Context) (bool, error) {
	return v.commentable().ViewerCanComment(ctx)
}

func (v ToCommentable) ViewerCannotCommentReasons(ctx context.Context) ([]CannotCommentReason, error) {
	return v.commentable().ViewerCannotCommentReasons(ctx)
}

func (v ToCommentable) Comments(ctx context.Context, arg *graphqlutil.ConnectionArgs) (CommentConnection, error) {
	return v.commentable().Comments(ctx, arg)
}

func (v ToCommentable) ToCampaign() (Campaign, bool) { return v.Campaign, v.Campaign != nil }
func (v ToCommentable) ToThread() (Thread, bool)     { return v.Thread, v.Thread != nil }
