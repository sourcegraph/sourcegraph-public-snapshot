package backend

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

var Users = &users{}

type users struct{}

func (s *users) List(ctx context.Context) (res *sourcegraph.UserList, err error) {
	if Mocks.Users.List != nil {
		return Mocks.Users.List(ctx)
	}

	// 🚨 SECURITY:  only admins are allowed to use this endpoint
	if err := CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	users, err := localstore.Users.List(ctx)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.UserList{Users: users}, nil
}
