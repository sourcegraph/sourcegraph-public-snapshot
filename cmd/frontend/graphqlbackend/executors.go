package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

const DefaultExecutorsLimit = 50

func (r *schemaResolver) Executors(ctx context.Context, args *struct {
	Query  *string
	Active *bool
	First  *int32
	After  *string
}) (*executorConnectionResolver, error) {
	// 🚨 SECURITY: Only site-admins may view executor details
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	query := ""
	if args.Query != nil {
		query = *args.Query
	}

	active := false
	if args.Active != nil {
		active = *args.Active
	}

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return nil, err
	}

	limit := DefaultExecutorsLimit
	if args.First != nil {
		limit = int(*args.First)
	}

	executors, totalCount, err := r.db.Executors().List(ctx, database.ExecutorStoreListOptions{
		Query:  query,
		Active: active,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*ExecutorResolver, 0, len(executors))
	for _, executor := range executors {
		resolvers = append(resolvers, &ExecutorResolver{executor: executor})
	}

	return &executorConnectionResolver{
		resolvers:  resolvers,
		totalCount: totalCount,
		nextOffset: graphqlutil.NextOffset(offset, len(executors), totalCount),
	}, nil
}

func executorByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*ExecutorResolver, error) {
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

	return &ExecutorResolver{executor: executor}, nil
}
