package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

func threadByID(ctx context.Context, id graphql.ID) (*threadResolver, error) {
	threadID, err := unmarshalThreadID(id)
	if err != nil {
		return nil, err
	}
	return threadByIDInt32(ctx, threadID)
}

func threadByIDInt32(ctx context.Context, threadID int32) (*threadResolver, error) {
	thread, err := db.Threads.Get(ctx, threadID)
	if err != nil {
		return nil, err
	}

	repo, err := db.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return nil, err
	}

	org, err := db.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is a member of the org.
	if err := backend.CheckOrgAccess(ctx, org.ID); err != nil {
		return nil, err
	}

	return &threadResolver{
		thread: thread,
		repo:   repo,
		org:    org,
	}, nil
}

func marshalThreadID(id int32) graphql.ID { return relay.MarshalID("Thread", id) }

func unmarshalThreadID(id graphql.ID) (threadID int32, err error) {
	err = relay.UnmarshalSpec(id, &threadID)
	return
}

func (r *threadResolver) ID() graphql.ID { return relay.MarshalID("Thread", r.thread.ID) }

func (r *threadResolver) DatabaseID() int32 { return r.thread.ID }
