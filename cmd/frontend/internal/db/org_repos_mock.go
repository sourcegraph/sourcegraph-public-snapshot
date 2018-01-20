package db

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"

	"context"
)

type MockOrgRepos struct {
	GetByID                func(ctx context.Context, id int32) (*types.OrgRepo, error)
	GetByCanonicalRemoteID func(ctx context.Context, orgID int32, CanonicalRemoteID string) (*types.OrgRepo, error)
	Create                 func(ctx context.Context, newRepo *types.OrgRepo) (*types.OrgRepo, error)
}

func (s *MockOrgRepos) MockGetByID_Return(t *testing.T, returns *types.OrgRepo, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByID = func(ctx context.Context, id int32) (*types.OrgRepo, error) {
		*called = true
		return returns, returnsErr
	}
	return
}

func (s *MockOrgRepos) MockGetByCanonicalRemoteID_Return(t *testing.T, returns *types.OrgRepo, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByCanonicalRemoteID = func(ctx context.Context, orgID int32, CanonicalRemoteID string) (*types.OrgRepo, error) {
		*called = true
		return returns, returnsErr
	}
	return
}

func (s *MockOrgRepos) MockCreate_Return(t *testing.T, returns *types.OrgRepo, returnsErr error) (called *bool, calledWith *types.OrgRepo) {
	called, calledWith = new(bool), &types.OrgRepo{}
	s.Create = func(ctx context.Context, newRepo *types.OrgRepo) (*types.OrgRepo, error) {
		*called, *calledWith = true, *newRepo
		return returns, returnsErr
	}
	return called, calledWith
}
