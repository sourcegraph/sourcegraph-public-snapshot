package backend

import (
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockUsers struct {
	List func(context.Context) (*sourcegraph.UserList, error)
}

func (s *MockUsers) MockList(t *testing.T, wantUsernames ...string) (called *bool) {
	called = new(bool)
	s.List = func(ctx context.Context) (*sourcegraph.UserList, error) {
		*called = true
		users := make([]*sourcegraph.User, len(wantUsernames))
		for i, username := range wantUsernames {
			users[i] = &sourcegraph.User{Username: username}
		}
		return &sourcegraph.UserList{Users: users}, nil
	}
	return
}
