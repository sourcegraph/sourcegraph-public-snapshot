package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

func (r *siteResolver) Orgs(args *struct {
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
