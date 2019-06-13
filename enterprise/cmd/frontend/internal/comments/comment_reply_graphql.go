package comments

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/internal"
)

func (GraphQLResolver) CommentReplyByID(ctx context.Context, id graphql.ID) (graphqlbackend.CommentReply, error) {
	dbID, err := unmarshalCommentReplyID(id)
	if err != nil {
		return nil, err
	}
	return CommentReplyByDBID(ctx, dbID)
}

func CommentReplyByDBID(ctx context.Context, dbID int64) (graphqlbackend.CommentReply, error) {
	v, err := internal.DBComments{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlCommentReply{gqlComment: &gqlComment{dbComment: v}}, nil
}

const gqlTypeCommentReply = "CommentReply"

type gqlCommentReply struct {
	*gqlComment
}

func (v *gqlCommentReply) ID() graphql.ID {
	return marshalCommentReplyID(v.gqlComment.dbComment.ID)
}

func marshalCommentReplyID(id int64) graphql.ID {
	return relay.MarshalID(gqlTypeCommentReply, id)
}

func unmarshalCommentReplyID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func (v *gqlCommentReply) Parent(ctx context.Context) (graphqlbackend.Comment, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return nil, err
	}
	if c.Object.ParentCommentID == 0 {
		return nil, nil
	}
	dbParentComment, err := internal.DBComments{}.GetByID(ctx, c.Object.ParentCommentID)
	if err != nil {
		return nil, err
	}
	return newGQLToComment(ctx, dbParentComment)
}

func (v *gqlCommentReply) ViewerCanUpdate(ctx context.Context) (bool, error) {
	return commentobjectdb.ViewerCanUpdate(ctx, v.ID())
}

func (v *gqlCommentReply) ViewerCanComment(ctx context.Context) (bool, error) {
	return commentobjectdb.ViewerCanComment(ctx)
}

func (v *gqlCommentReply) ViewerCannotCommentReasons(ctx context.Context) ([]graphqlbackend.CannotCommentReason, error) {
	return commentobjectdb.ViewerCannotCommentReasons(ctx)
}

func (v *gqlCommentReply) Comments(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.CommentConnection, error) {
	return graphqlbackend.CommentsForObject(ctx, v.ID(), arg)
}

func (gqlCommentReply) IsCommentReply() {}
