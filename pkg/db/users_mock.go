package db

import (
	"context"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockUsers struct {
	Create               func(ctx context.Context, info NewUser) (newUser *sourcegraph.User, err error)
	GetByID              func(ctx context.Context, id int32) (*sourcegraph.User, error)
	GetByUsername        func(ctx context.Context, username string) (*sourcegraph.User, error)
	GetByExternalID      func(ctx context.Context, provider, id string) (*sourcegraph.User, error)
	GetByCurrentAuthUser func(ctx context.Context) (*sourcegraph.User, error)
	Count                func(ctx context.Context) (int, error)
	ListByOrg            func(ctx context.Context, orgID int32, userIDs []int32, usernames []string) ([]*sourcegraph.User, error)
}

func (s *MockUsers) MockGetByID_Return(t *testing.T, returns *sourcegraph.User, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByID = func(ctx context.Context, id int32) (*sourcegraph.User, error) {
		*called = true
		return returns, returnsErr
	}
	return
}

func (s *MockUsers) MockGetByExternalID_Return(t *testing.T, returns *sourcegraph.User, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByExternalID = func(ctx context.Context, provider, id string) (*sourcegraph.User, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
