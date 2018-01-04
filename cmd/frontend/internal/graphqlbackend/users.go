package graphqlbackend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func (r *siteResolver) Users(args *struct {
	connectionArgs
	Query *string
}) *userConnectionResolver {
	return &userConnectionResolver{
		connectionResolverCommon: newConnectionResolverCommon(args.connectionArgs),
		query: args.Query,
	}
}

type userConnectionResolver struct {
	connectionResolverCommon
	query *string
}

func (r *userConnectionResolver) Nodes(ctx context.Context) ([]*userResolver, error) {
	// 🚨 SECURITY: Only site admins can list users.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	opt := &db.UsersListOptions{
		ListOptions: sourcegraph.ListOptions{
			PerPage: r.first,
		},
	}
	if r.query != nil {
		opt.Query = *r.query
	}

	users, err := db.Users.List(ctx, opt)
	if err != nil {
		return nil, err
	}

	var l []*userResolver
	for _, user := range users {
		l = append(l, &userResolver{
			user: user,
		})
	}
	return l, nil
}

func (r *userConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can count users.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return 0, err
	}

	count, err := db.Users.Count(ctx)
	return int32(count), err
}
