package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
)

// CommentReplyByID is called to look up a CommentReply given its GraphQL ID.
func CommentReplyByID(ctx context.Context, id graphql.ID) (CommentReply, error) {
	if Comments == nil {
		return nil, errors.New("comments is not implemented")
	}
	return Comments.CommentReplyByID(ctx, id)
}

// CommentReply is the interface for the GraphQL type CommentReply.
type CommentReply interface {
	ID() graphql.ID
	PartialComment
	Updatable
	Parent(context.Context) (Comment, error)

	// IsCommentReply is a tag to distinguish this interface from other interfaces that are
	// otherwise a superset of it, such as Campaign. If Campaign implements CommentReply, the
	// (*NodeResolver).ToXyz methods get confused and report a Campaign's __typename as
	// CommentReply.
	//
	// TODO!(sqs): maybe not necessary now that this interface has the Parent method, which other
	// Comment implementations don't have.
	IsCommentReply()
}
