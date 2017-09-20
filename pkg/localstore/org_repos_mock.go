package localstore

import (
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockOrgRepos struct {
	GetByAccessToken func(ctx context.Context, remoteURI, accessToken string) (*sourcegraph.OrgRepo, error)
	Create           func(ctx context.Context, newRepo *sourcegraph.OrgRepo) (*sourcegraph.OrgRepo, error)
}

func (s *MockOrgRepos) MockGet_Return(t *testing.T, returns *sourcegraph.OrgRepo, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByAccessToken = func(ctx context.Context, remoteURI, accessToken string) (*sourcegraph.OrgRepo, error) {
		*called = true
		return returns, returnsErr
	}
	return
}

func (s *MockOrgRepos) MockCreate_Return(t *testing.T, returns *sourcegraph.OrgRepo, returnsErr error) (called *bool, calledWith *sourcegraph.OrgRepo) {
	called, calledWith = new(bool), &sourcegraph.OrgRepo{}
	s.Create = func(ctx context.Context, newRepo *sourcegraph.OrgRepo) (*sourcegraph.OrgRepo, error) {
		*called, *calledWith = true, *newRepo
		return returns, returnsErr
	}
	return called, calledWith
}
