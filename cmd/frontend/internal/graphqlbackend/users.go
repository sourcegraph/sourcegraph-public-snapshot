package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

func (r *siteResolver) Users(args *struct {
	connectionArgs
	Query *string
}) *userConnectionResolver {
	var opt db.UsersListOptions
	if args.Query != nil {
		opt.Query = *args.Query
	}
	args.connectionArgs.set(&opt.LimitOffset)
	return &userConnectionResolver{opt: opt}
}

type userConnectionResolver struct {
	opt db.UsersListOptions
}

func (r *userConnectionResolver) Nodes(ctx context.Context) ([]*userResolver, error) {
	// 🚨 SECURITY: Only site admins can list users.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	users, err := db.Users.List(ctx, &r.opt)
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

	count, err := db.Users.Count(ctx, r.opt)
	return int32(count), err
}

// staticUserConnectionResolver implements the GraphQL type UserConnection based on an underlying
// list of users that is computed statically.
type staticUserConnectionResolver struct {
	users []*types.User
}

func (r *staticUserConnectionResolver) Nodes() []*userResolver {
	resolvers := make([]*userResolver, len(r.users))
	for i, user := range r.users {
		resolvers[i] = &userResolver{user: user}
	}
	return resolvers
}

func (r *staticUserConnectionResolver) TotalCount() int32 { return int32(len(r.users)) }
