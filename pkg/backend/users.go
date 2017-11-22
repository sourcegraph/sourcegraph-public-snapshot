package backend

import (
	"context"

	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

var Users = &users{}

type users struct{}

func (s *users) List(ctx context.Context) (res *sourcegraph.UserList, err error) {
	actor := actor.FromContext(ctx)
	// 🚨 SECURITY:  only admins are allowed to use this endpoint
	if !actor.IsAdmin() {
		return nil, errors.New("Must be an admin")
	}
	users, err := localstore.Users.List(ctx)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.UserList{Users: users}, nil
}
