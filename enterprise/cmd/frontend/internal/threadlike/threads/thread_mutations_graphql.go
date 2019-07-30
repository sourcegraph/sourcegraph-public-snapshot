package threads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func (GraphQLResolver) CreateThread(ctx context.Context, arg *graphqlbackend.CreateThreadArgs) (graphqlbackend.Thread, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	thread, err := internal.DBThreads{}.Create(ctx, &internal.DBThread{
		Type:         graphqlbackend.ThreadlikeTypeThread,
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		ExternalURL:  arg.Input.ExternalURL,
		Status:       string(graphqlbackend.ThreadStatusOpen),
	})
	if err != nil {
		return nil, err
	}
	return newGQLThread(thread), nil
}

func (GraphQLResolver) UpdateThread(ctx context.Context, arg *graphqlbackend.UpdateThreadArgs) (graphqlbackend.Thread, error) {
	l, err := threadByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	thread, err := internal.DBThreads{}.Update(ctx, l.db.ID, internal.DBThreadUpdate{
		Title: arg.Input.Title,
		// TODO!(sqs): handle body update
		ExternalURL: arg.Input.ExternalURL,
	})
	if err != nil {
		return nil, err
	}
	return newGQLThread(thread), nil
}

func (GraphQLResolver) DeleteThread(ctx context.Context, arg *graphqlbackend.DeleteThreadArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlThread, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	return nil, internal.DBThreads{}.DeleteByID(ctx, gqlThread.db.ID)
}
