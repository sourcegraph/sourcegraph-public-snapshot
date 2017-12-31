package db

import (
	"context"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockUsers struct {
	Create               func(ctx context.Context, authID, email, username, displayName, provider string, avatarURL *string, password string, emailCode string) (newUser *sourcegraph.User, err error)
	GetByID              func(ctx context.Context, id int32) (*sourcegraph.User, error)
	GetByUsername        func(ctx context.Context, username string) (*sourcegraph.User, error)
	GetByAuthID          func(ctx context.Context, id string) (*sourcegraph.User, error)
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

func (s *MockUsers) MockGetByAuthID_Return(t *testing.T, returns *sourcegraph.User, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByAuthID = func(ctx context.Context, id string) (*sourcegraph.User, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
