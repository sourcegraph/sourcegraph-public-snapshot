package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func unmarshalExecutorID(id graphql.ID) (executorID int64, err error) {
	err = relay.UnmarshalSpec(id, &executorID)
	return
}

type ExecutorsListArgs struct {
	Query  *string
	Active *bool
	First  int32
	After  *string
}

func (r *schemaResolver) Executors(ctx context.Context, args ExecutorsListArgs) (*executorConnectionResolver, error) {
	// 🚨 SECURITY: Only site-admins may view executor details
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	offset, err := gqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return nil, err
	}

	var executorConnection *executorConnectionResolver
	err = r.db.WithTransact(ctx, func(tx database.DB) error {
		opts := database.ExecutorStoreListOptions{
			Offset: offset,
			Limit:  int(args.First),
		}
		if args.Query != nil {
			opts.Query = *args.Query
		}
		if args.Active != nil {
			opts.Active = *args.Active
		}
		execs, err := tx.Executors().List(ctx, opts)
		if err != nil {
			return err
		}
		totalCount, err := tx.Executors().Count(ctx, opts)
		if err != nil {
			return err
		}

		resolvers := make([]*ExecutorResolver, 0, len(execs))
		for _, executor := range execs {
			resolvers = append(resolvers, &ExecutorResolver{executor: executor})
		}

		nextOffset := gqlutil.NextOffset(offset, len(execs), totalCount)

		executorConnection = &executorConnectionResolver{
			resolvers:  resolvers,
			totalCount: totalCount,
			nextOffset: nextOffset,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return executorConnection, nil
}

func (r *schemaResolver) AreExecutorsConfigured() bool {
	return conf.ExecutorsAccessToken() != ""
}

func executorByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*ExecutorResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
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

	return NewExecutorResolver(executor), nil
}
