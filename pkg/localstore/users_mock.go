package localstore

import (
	"context"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockUsers struct {
	GetByID              func(ctx context.Context, id int32) (*sourcegraph.User, error)
	GetByAuth0ID         func(ctx context.Context, id string) (*sourcegraph.User, error)
	GetByCurrentAuthUser func(ctx context.Context) (*sourcegraph.User, error)
	ListByOrg            func(ctx context.Context, orgID int32, auth0IDs, usernames []string) ([]*sourcegraph.User, error)
}

func (s *MockUsers) MockGetByID_Return(t *testing.T, returns *sourcegraph.User, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByID = func(ctx context.Context, id int32) (*sourcegraph.User, error) {
		*called = true
		return returns, returnsErr
	}
	return
}

func (s *MockUsers) MockGetByAuth0ID_Return(t *testing.T, returns *sourcegraph.User, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByAuth0ID = func(ctx context.Context, id string) (*sourcegraph.User, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
