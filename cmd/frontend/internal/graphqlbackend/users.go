package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func (r *schemaResolver) Users(args *struct {
	connectionArgs
}) *userConnectionResolver {
	return &userConnectionResolver{
		connectionResolverCommon: newConnectionResolverCommon(args.connectionArgs),
	}
}

type userConnectionResolver struct {
	connectionResolverCommon
}

func (r *userConnectionResolver) Nodes(ctx context.Context) ([]*userResolver, error) {
	// 🚨 SECURITY: Only site admins can list users.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	usersList, err := backend.Users.List(ctx)
	if err != nil {
		return nil, err
	}

	var l []*userResolver
	for _, user := range usersList.Users {
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

	count, err := localstore.Users.Count(ctx)
	return int32(count), err
}
