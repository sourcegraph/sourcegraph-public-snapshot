package localstore

import (
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockLocalRepos struct {
	Get    func(ctx context.Context, remoteURI, accessToken string, orgID int32) (*sourcegraph.LocalRepo, error)
	Create func(ctx context.Context, newRepo *sourcegraph.LocalRepo) (*sourcegraph.LocalRepo, error)
}

func (s *MockLocalRepos) MockGet_Return(t *testing.T, returns *sourcegraph.LocalRepo, returnsErr error) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, remoteURI, accessToken string, orgID int32) (*sourcegraph.LocalRepo, error) {
		*called = true
		return returns, returnsErr
	}
	return
}

func (s *MockLocalRepos) MockCreate_Return(t *testing.T, returns *sourcegraph.LocalRepo, returnsErr error) (called *bool, calledWith *sourcegraph.LocalRepo) {
	called, calledWith = new(bool), &sourcegraph.LocalRepo{}
	s.Create = func(ctx context.Context, newRepo *sourcegraph.LocalRepo) (*sourcegraph.LocalRepo, error) {
		*called, *calledWith = true, *newRepo
		return returns, returnsErr
	}
	return called, calledWith
}
