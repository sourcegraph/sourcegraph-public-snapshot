package commentobjectdb

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func ViewerCanUpdate(ctx context.Context, id graphql.ID) (bool, error) {
	return true, nil
	// TODO!(sqs): commented out below because commentByGQLID is not in this package
	//
	// // TODO!(sqs): add tests, verify this is desired behavior
	// v, err := commentByGQLID(ctx, id)
	// if err != nil {
	// 	return false, err
	// }
	// err = backend.CheckSiteAdminOrSameUser(ctx, v.AuthorUserID)
	// return err == nil, err
}

func ViewerCanComment(ctx context.Context) (bool, error) {
	// TODO!(sqs): add tests, verify this is desired behavior
	actor, err := graphqlbackend.CurrentUser(ctx)
	return actor != nil && err == nil, err
}

func ViewerCannotCommentReasons(ctx context.Context) ([]graphqlbackend.CannotCommentReason, error) {
	actor, err := graphqlbackend.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if actor == nil {
		return []graphqlbackend.CannotCommentReason{graphqlbackend.CannotCommentReasonAuthenticationRequired}, nil
	}
	return nil, nil
}
