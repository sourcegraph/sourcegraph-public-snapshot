package db

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"

	"context"
)

type MockThreads struct {
	Get    func(ctx context.Context, id int32) (*types.Thread, error)
	Create func(ctx context.Context, newThread *types.Thread) (*types.Thread, error)
	Update func(ctx context.Context, id int32, repo api.RepoID, archived *bool) (*types.Thread, error)
	List   func(ctx context.Context, opt *ThreadsListOptions) ([]*types.Thread, error)
	Count  func(ctx context.Context, opt ThreadsListOptions) (int, error)
}

func (s *MockThreads) MockGet_Return(t *testing.T, returns *types.Thread, returnsErr error) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, id int32) (*types.Thread, error) {
		*called = true
		return returns, returnsErr
	}
	return called
}

func (s *MockThreads) MockCreate_Return(t *testing.T, returns *types.Thread, returnsErr error) (called *bool, calledWith *types.Thread) {
	called, calledWith = new(bool), &types.Thread{}
	s.Create = func(ctx context.Context, newThread *types.Thread) (*types.Thread, error) {
		*called = true
		return returns, returnsErr
	}
	return called, calledWith
}

func (s *MockThreads) MockUpdate_Return(t *testing.T, returns *types.Thread, returnsErr error) (called *bool) {
	called = new(bool)
	s.Update = func(ctx context.Context, id int32, repo api.RepoID, archived *bool) (*types.Thread, error) {
		*called = true
		return returns, returnsErr
	}
	return called
}
