package comments

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/internal"
)

func CommentActorFromContext(ctx context.Context) (actor.DBColumns, error) {
	user, err := graphqlbackend.CurrentUser(ctx)
	if err != nil {
		return actor.DBColumns{}, err
	}
	if user == nil {
		return actor.DBColumns{}, errors.New("authenticated required to create comment")
	}
	return actor.DBColumns{UserID: user.DatabaseID()}, nil
}

type ExternalComment struct {
	commentobjectdb.DBObjectCommentFields
}

func (GraphQLResolver) EditComment(ctx context.Context, arg *graphqlbackend.EditCommentArgs) (graphqlbackend.Comment, error) {
	v, err := commentByGQLID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	comment, err := internal.DBComments{}.Update(ctx, v.ID, internal.DBCommentUpdate{
		Body: &arg.Input.Body,
	})
	if err != nil {
		return nil, err
	}
	return newGQLToComment(ctx, comment)
}

func (GraphQLResolver) DeleteComment(ctx context.Context, arg *graphqlbackend.DeleteCommentArgs) (*graphqlbackend.EmptyResponse, error) {
	v, err := commentByGQLID(ctx, arg.Comment)
	if err != nil {
		return nil, err
	}
	return nil, internal.DBComments{}.DeleteByID(ctx, v.ID)
}
