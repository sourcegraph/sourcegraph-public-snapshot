package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type executorResolver struct {
	executor database.Executor
}

type executorConnectionResolver struct {
	resolvers []*executorResolver
}

func (r *schemaResolver) Executors(ctx context.Context, args *struct {
	First *int32
	After *string
}) (*executorConnectionResolver, error) {
	executors, _, err := r.db.Executors().List(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*executorResolver, 0, len(executors))
	for _, executor := range executors {
		resolvers = append(resolvers, &executorResolver{executor: executor})
	}

	return &executorConnectionResolver{
		resolvers: resolvers,
	}, nil
}

func (r *executorConnectionResolver) Nodes(ctx context.Context) []*executorResolver {
	return r.resolvers
}

func (r *executorConnectionResolver) TotalCount(ctx context.Context) int32 {
	return int32(len(r.resolvers))
}

func (r *executorConnectionResolver) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
	return nil
}

func marshalExecutorID(id int64) graphql.ID {
	return relay.MarshalID("Executor", id)
}

func unmarshalExecutorID(id graphql.ID) (executorID int64, err error) {
	err = relay.UnmarshalSpec(id, &executorID)
	return
}

func executorByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*executorResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmarshalExecutorID(gqlID)
	if err != nil {
		return nil, err
	}

	executor, ok, err := db.Executors().GetByID(ctx, int(id))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &executorResolver{executor: executor}, nil
}

func (e *executorResolver) ID() graphql.ID {
	return marshalExecutorID(int64(e.executor.ID))
}

func (e *executorResolver) Hostname() string {
	return e.executor.Hostname
}

func (e *executorResolver) LastSeenAt() DateTime {
	return DateTime{e.executor.LastSeenAt}
}
