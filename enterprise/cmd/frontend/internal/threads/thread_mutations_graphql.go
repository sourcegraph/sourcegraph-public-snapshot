package threads

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) UpdateThread(ctx context.Context, arg *graphqlbackend.UpdateThreadArgs) (graphqlbackend.Thread, error) {
	l, err := threadByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	thread, err := dbThreads{}.Update(ctx, l.db.ID, dbThreadUpdate{
		Title: arg.Input.Title,
		// TODO!(sqs): handle body update
		BaseRef: arg.Input.BaseRef,
		HeadRef: arg.Input.HeadRef,
	})
	if err != nil {
		return nil, err
	}
	return newGQLThread(thread), nil
}

func (GraphQLResolver) MarkThreadAsReady(ctx context.Context, arg *graphqlbackend.MarkThreadAsReadyArgs) (graphqlbackend.Thread, error) {
	t, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	if !t.db.IsDraft {
		return nil, errors.New("thread is not a draft thread")
	}
	tmp := false
	thread, err := dbThreads{}.Update(ctx, t.db.ID, dbThreadUpdate{
		IsDraft: &tmp,
	})
	if err != nil {
		return nil, err
	}
	return newGQLThread(thread), nil
}

func (GraphQLResolver) PublishThreadToExternalService(ctx context.Context, arg *graphqlbackend.PublishThreadToExternalServiceArgs) (graphqlbackend.Thread, error) {
	t, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	if err := PublishThreadToExternalService(ctx, t); err != nil {
		return nil, err
	}
	return threadByID(ctx, arg.Thread)
}

func PublishThreadToExternalService(ctx context.Context, thread_ graphqlbackend.Thread) error {
	thread := thread_.(*gqlThread).db
	if !thread.IsPendingExternalCreation {
		return errors.New("thread is not pending external creation")
	}

	repo, err := thread_.Repository(ctx)
	if err != nil {
		return err
	}
	threadBody, err := thread_.Body(ctx)
	if err != nil {
		return err
	}

	// TODO!(sqs): use actual campaign name, but its weird because a thread can be in >1 campaigns
	campaignName := thread_.Title()
	_, err = CreateOnExternalService(ctx, thread.ID, thread_.Title(), threadBody, campaignName, repo, []byte(thread.PendingPatch))
	return err
}

func (GraphQLResolver) DeleteThread(ctx context.Context, arg *graphqlbackend.DeleteThreadArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlThread, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	return nil, dbThreads{}.DeleteByID(ctx, gqlThread.db.ID)
}
