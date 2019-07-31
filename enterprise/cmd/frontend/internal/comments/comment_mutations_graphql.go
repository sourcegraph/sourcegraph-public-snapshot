package comments

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateComment(ctx context.Context, arg *graphqlbackend.CreateCommentArgs) (graphqlbackend.Comment, error) {
	// TODO!(sqs): add auth checks
	actor, err := graphqlbackend.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if actor == nil {
		return nil, errors.New("authenticated required to create comment")
	}

	v := &DBComment{
		AuthorUserID: actor.DatabaseID(),
		Body:         arg.Input.Body,
	}
	v.ThreadID, err = commentLookupInfoFromGQLID(arg.Input.Node)
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
