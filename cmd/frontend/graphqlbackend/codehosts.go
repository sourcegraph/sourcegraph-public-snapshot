package graphqlbackend

import (
	"context"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (r *schemaResolver) Codehosts(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*codehostConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read codehost (they have secrets).
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	var opt db.CodehostsListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &codehostConnectionResolver{opt: opt}, nil
}

type codehostConnectionResolver struct {
	opt db.CodehostsListOptions

	// cache results because they are used by multiple fields
	once      sync.Once
	codehosts []*types.Codehost
	err       error
}

func (r *codehostConnectionResolver) compute(ctx context.Context) ([]*types.Codehost, error) {
	r.once.Do(func() {
		r.codehosts, r.err = db.Codehosts.List(ctx, r.opt)
	})
	return r.codehosts, r.err
}

func (r *codehostConnectionResolver) Nodes(ctx context.Context) ([]*codehostResolver, error) {
	codehosts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*codehostResolver, 0, len(codehosts))
	for _, codehost := range codehosts {
		resolvers = append(resolvers, &codehostResolver{codehost: codehost})
	}
	return resolvers, nil
}

func (r *codehostConnectionResolver) TotalCount(ctx context.Context) (countptr int32, err error) {
	count, err := db.Codehosts.Count(ctx, r.opt)
	return int32(count), err
}

func (r *codehostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	codehosts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(codehosts) >= r.opt.Limit), nil
}
