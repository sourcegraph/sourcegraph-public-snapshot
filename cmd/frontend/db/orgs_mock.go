package db

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"

	"context"
)

type MockOrgs struct {
	GetByID   func(ctx context.Context, id int32) (*types.Org, error)
	GetByName func(ctx context.Context, name string) (*types.Org, error)
	Count     func(ctx context.Context, opt OrgsListOptions) (int, error)
	List      func(ctx context.Context, opt *OrgsListOptions) ([]*types.Org, error)
}

func (s *MockOrgs) MockGetByID_Return(t *testing.T, returns *types.Org, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByID = func(ctx context.Context, id int32) (*types.Org, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
