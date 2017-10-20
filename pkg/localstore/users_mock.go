package localstore

import (
	"context"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockUsers struct {
	GetByAuth0ID func(id string) (*sourcegraph.User, error)
	ListByOrg    func(ctx context.Context, orgID int32, auth0IDs, usernames []string) ([]*sourcegraph.User, error)
}

func (s *MockUsers) MockGetByAuth0ID_Return(t *testing.T, returns *sourcegraph.User, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByAuth0ID = func(id string) (*sourcegraph.User, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
