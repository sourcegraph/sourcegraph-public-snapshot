package graphql

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	livedependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/live"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type Resolver interface {
	LockfileIndexes(ctx context.Context, args *graphqlbackend.ListLockfileIndexesArgs) (graphqlbackend.LockfileIndexConnectionResolver, error)
}

type resolver struct {
	svc *dependencies.Service
	db  database.DB
}

func New(db database.DB) Resolver {
	return &resolver{
		svc: livedependencies.GetService(db, livedependencies.NewSyncer()),
		db:  db,
	}
}

func (r *resolver) LockfileIndexes(ctx context.Context, args *graphqlbackend.ListLockfileIndexesArgs) (graphqlbackend.LockfileIndexConnectionResolver, error) {
	// 🚨 SECURITY: For now we only allow site admins to query lockfile indexes.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	p, err := validateArgs(args)
	if err != nil {
		return nil, err
	}

	opts := dependencies.ListLockfileIndexesOpts{
		After: p.after,
		Limit: p.limit,
	}

	lockfileIndexes, totalCount, err := r.svc.ListLockfileIndexes(ctx, opts)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*LockfileIndexResolver, 0, len(lockfileIndexes))
	for _, executor := range lockfileIndexes {
		resolvers = append(resolvers, NewExecutorResolver(executor))
	}

	nextOffset := graphqlutil.NextOffset(p.after, len(lockfileIndexes), totalCount)
	lockfileIndexesConnection := NewLockfileIndexConnectionConnection(resolvers, totalCount, nextOffset)

	return lockfileIndexesConnection, nil
}

const DefaultLockfileIndexesLimit = 50

type params struct {
	after int
	limit int
}

func validateArgs(args *graphqlbackend.ListLockfileIndexesArgs) (params, error) {
	var p params
	afterCount, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return p, err
	}
	p.after = afterCount

	limit := DefaultLockfileIndexesLimit
	if args.First != 0 {
		limit = int(args.First)
	}
	p.limit = limit

	return p, nil
}
