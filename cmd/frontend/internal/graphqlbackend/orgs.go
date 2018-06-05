package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

func (r *schemaResolver) Organizations(args *struct {
	connectionArgs
	Query *string
}) *orgConnectionResolver {
	var opt db.OrgsListOptions
	if args.Query != nil {
		opt.Query = *args.Query
	}
	args.connectionArgs.set(&opt.LimitOffset)
	return &orgConnectionResolver{opt: opt}
}

type orgConnectionResolver struct {
	opt db.OrgsListOptions
}

func (r *orgConnectionResolver) Nodes(ctx context.Context) ([]*orgResolver, error) {
	// 🚨 SECURITY: Only site admins can list orgs.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	orgs, err := db.Orgs.List(ctx, &r.opt)
	if err != nil {
		return nil, err
	}

	var l []*orgResolver
	for _, org := range orgs {
		l = append(l, &orgResolver{
			org: org,
		})
	}
	return l, nil
}

func (r *orgConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can count orgs.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return 0, err
	}

	count, err := db.Orgs.Count(ctx, r.opt)
	return int32(count), err
}

type orgConnectionStaticResolver struct {
	nodes []*orgResolver
}

func (r *orgConnectionStaticResolver) Nodes() []*orgResolver { return r.nodes }
func (r *orgConnectionStaticResolver) TotalCount() int32     { return int32(len(r.nodes)) }
func (r *orgConnectionStaticResolver) PageInfo() *pageInfo   { return &pageInfo{hasNextPage: false} }
