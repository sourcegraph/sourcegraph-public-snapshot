package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

func (r *schemaResolver) Users(ctx context.Context) (*userConnectionResolver, error) {
	users, err := listUsers(ctx)
	if err != nil {
		return nil, err
	}
	return &userConnectionResolver{users: users}, nil
}

func listUsers(ctx context.Context) ([]*userResolver, error) {
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

type userConnectionResolver struct {
	users []*userResolver
}

func (r *userConnectionResolver) Nodes() []*userResolver { return r.users }

func (r *userConnectionResolver) TotalCount() int32 { return int32(len(r.users)) }
