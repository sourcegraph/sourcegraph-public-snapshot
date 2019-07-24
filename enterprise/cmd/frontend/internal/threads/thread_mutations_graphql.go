package threads

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateThread(ctx context.Context, arg *graphqlbackend.CreateThreadArgs) (graphqlbackend.Thread, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	db := &dbThread{
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		ExternalURL:  arg.Input.ExternalURL,
		Type:         arg.Input.Type,
	}
	// Apply default status.
	if arg.Input.Status != nil {
		db.Status = *arg.Input.Status
	} else {
		db.Status = graphqlbackend.ThreadStatusOpen
	}

	// Validate.
	if !graphqlbackend.IsValidThreadStatus(string(db.Status)) {
		return nil, errors.New("invalid thread status")
	}
	if !graphqlbackend.IsValidThreadType(string(db.Type)) {
		return nil, errors.New("invalid thread type")
	}

	thread, err := dbThreads{}.Create(ctx, db)
	if err != nil {
		return nil, err
	}
	return &gqlThread{db: thread}, nil
}

func (GraphQLResolver) UpdateThread(ctx context.Context, arg *graphqlbackend.UpdateThreadArgs) (graphqlbackend.Thread, error) {
	update := dbThreadUpdate{
		Title:       arg.Input.Title,
		ExternalURL: arg.Input.ExternalURL,
		Status:      arg.Input.Status,
	}
	if update.Status != nil && !graphqlbackend.IsValidThreadStatus(string(*update.Status)) {
		return nil, errors.New("invalid thread status")
	}

	l, err := threadByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	thread, err := dbThreads{}.Update(ctx, l.db.ID, update)
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
