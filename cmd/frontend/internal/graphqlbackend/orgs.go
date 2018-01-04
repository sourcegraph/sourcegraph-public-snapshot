package graphqlbackend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func (r *siteResolver) Orgs(args *struct {
	connectionArgs
	Query *string
}) *orgConnectionResolver {
	return &orgConnectionResolver{
		connectionResolverCommon: newConnectionResolverCommon(args.connectionArgs),
		query: args.Query,
	}
}

type orgConnectionResolver struct {
	connectionResolverCommon
	query *string
}

func (r *orgConnectionResolver) Nodes(ctx context.Context) ([]*orgResolver, error) {
	// 🚨 SECURITY: Only site admins can list orgs.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	opt := &db.OrgsListOptions{
		ListOptions: sourcegraph.ListOptions{
			PerPage: r.first,
		},
	}
	if r.query != nil {
		opt.Query = *r.query
	}

	orgs, err := db.Orgs.List(ctx, opt)
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

	count, err := db.Orgs.Count(ctx)
	return int32(count), err
}
