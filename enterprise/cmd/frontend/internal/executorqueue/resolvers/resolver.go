package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/services/executors"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type executorsResolver struct {
	conf   conftypes.UnifiedWatchable
	db     database.DB
	stores map[string]store.Store
	svc    executors.Executor
}

func NewExecutorsResolver(
	conf conftypes.UnifiedWatchable,
	db database.DB,
	stores map[string]store.Store,
) graphqlbackend.ExecutorsResolver {
	return &executorsResolver{
		conf:   conf,
		db:     db,
		stores: stores,
		svc:    executors.New(db),
	}
}

func (r *executorsResolver) Executors(
	ctx context.Context, args *graphqlbackend.ExecutorsArgs,
) (graphqlbackend.ExecutorConnectionResolver, error) {
	p, err := validateExecutorsArgs(args)
	if err != nil {
		return nil, err
	}

	execs, totalCount, err := r.svc.List(ctx, p.query, p.active, p.after, p.first)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.ExecutorResolver, 0, len(execs))
	for _, executor := range execs {
		resolvers = append(resolvers, NewExecutorResolver(executor))
	}

	nextOffset := graphqlutil.NextOffset(p.after, len(execs), totalCount)

	return &executorConnectionResolver{
		resolvers:  resolvers,
		totalCount: totalCount,
		nextOffset: nextOffset,
	}, nil
}

func (r *executorsResolver) AreExecutorsConfigured() bool {
	return r.conf.SiteConfig().ExecutorsAccessToken != ""
}

func (r *executorsResolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		"Executor": func(ctx context.Context, gqlID graphql.ID) (graphqlbackend.Node, error) {
			id, err := unmarshalExecutorID(gqlID)
			if err != nil {
				return nil, err
			}

			executor, ok, err := r.svc.GetByID(ctx, int(id))
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, nil
			}

			return NewExecutorResolver(executor), nil
		},
	}
}

func unmarshalExecutorID(id graphql.ID) (executorID int64, err error) {
	err = relay.UnmarshalSpec(id, &executorID)
	return
}

const DefaultExecutorsLimit = 50

func validateExecutorsArgs(args *graphqlbackend.ExecutorsArgs) (p struct {
	query  string
	active bool
	after  int
	first  int
}, err error) {
	if args.Query != nil {
		p.query = *args.Query
	}

	if args.Active != nil {
		p.active = *args.Active
	}

	after, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return p, err
	}
	p.after = after

	first := DefaultExecutorsLimit
	if args.First != nil {
		first = int(*args.First)
	}
	p.first = first

	return
}
