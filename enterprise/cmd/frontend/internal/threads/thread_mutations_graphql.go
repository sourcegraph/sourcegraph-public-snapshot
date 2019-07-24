package threads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateThread(ctx context.Context, arg *graphqlbackend.CreateThreadArgs) (graphqlbackend.Thread, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	thread, err := dbThreads{}.Create(ctx, &dbThread{
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		ExternalURL:  arg.Input.ExternalURL,
	})
	if err != nil {
		return nil, err
	}
	return &gqlThread{db: thread}, nil
}

func (GraphQLResolver) UpdateThread(ctx context.Context, arg *graphqlbackend.UpdateThreadArgs) (graphqlbackend.Thread, error) {
	l, err := threadByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	thread, err := dbThreads{}.Update(ctx, l.db.ID, dbThreadUpdate{
		Title:       arg.Input.Title,
		ExternalURL: arg.Input.ExternalURL,
	})
	if err != nil {
		return nil, err
	}
	return &gqlThread{db: thread}, nil
}

func (GraphQLResolver) DeleteThread(ctx context.Context, arg *graphqlbackend.DeleteThreadArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlThread, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	return nil, dbThreads{}.DeleteByID(ctx, gqlThread.db.ID)
}
