package localstore

import (
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockOrgs struct {
	GetByID func(ctx context.Context, id int32) (*sourcegraph.Org, error)
}

func (s *MockOrgs) MockGetByID_Return(t *testing.T, returns *sourcegraph.Org, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByID = func(ctx context.Context, id int32) (*sourcegraph.Org, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
