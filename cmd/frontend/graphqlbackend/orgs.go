package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type OrganizationsArgs struct {
	graphqlutil.ConnectionArgs
	After *string
	Query *string
}

func (r *schemaResolver) Organizations(ctx context.Context, args OrganizationsArgs) (*orgConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins can list organisations.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var opt database.OrgsListOptions
	if args.Query != nil {
		opt.Query = *args.Query
	}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &orgConnectionResolver{db: r.db, opt: opt}, nil
}

type orgConnectionResolver struct {
	db  database.DB
	opt database.OrgsListOptions
}

func (r *orgConnectionResolver) Nodes(ctx context.Context) ([]*OrgResolver, error) {
	orgs, err := r.db.Orgs().List(ctx, r.opt)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*OrgResolver, len(orgs))
	for i, org := range orgs {
		resolvers[i] = &OrgResolver{
			db:  r.db,
			org: org,
		}
	}

	return resolvers, nil
}

func (r *orgConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.db.Orgs().Count(ctx, r.opt)
	return int32(count), err
}

func (r *orgConnectionResolver) PageInfo(ctx context.Context) (graphqlutil.PageInfo, error) {
	return *graphqlutil.HasNextPage(false), nil
}
