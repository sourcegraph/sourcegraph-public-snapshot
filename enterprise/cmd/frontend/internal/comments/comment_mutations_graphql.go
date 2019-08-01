package comments

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func CommentActorFromContext(ctx context.Context) (authorUserID int32, err error) {
	actor, err := graphqlbackend.CurrentUser(ctx)
	if err != nil {
		return 0, err
	}
	if actor == nil {
		return 0, errors.New("authenticated required to create comment")
	}
	return actor.DatabaseID(), nil
}

func (GraphQLResolver) CreateComment(ctx context.Context, arg *graphqlbackend.CreateCommentArgs) (graphqlbackend.Comment, error) {
	// TODO!(sqs): add auth checks
	authorUserID, err := CommentActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	v := &dbComment{
		AuthorUserID: authorUserID,
		Body:         arg.Input.Body,
	}
	v.Object, err = commentObjectFromGQLID(arg.Input.Node)
	if err != nil {
		return nil, err
	}

	comment, err := dbComments{}.Create(ctx, v)
	if err != nil {
		return nil, err
	}
	return newGQLToComment(ctx, comment)
}

func (GraphQLResolver) EditComment(ctx context.Context, arg *graphqlbackend.EditCommentArgs) (graphqlbackend.Comment, error) {
	v, err := commentByGQLID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	comment, err := dbComments{}.Update(ctx, v.ID, dbCommentUpdate{
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
	return nil, dbComments{}.DeleteByID(ctx, v.ID)
}
