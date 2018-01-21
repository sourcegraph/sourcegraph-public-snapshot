package db

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"

	"context"
)

type MockOrgMembers struct {
	GetByOrgIDAndUserID func(ctx context.Context, orgID, userID int32) (*types.OrgMember, error)
}

func (s *MockOrgMembers) MockGetByOrgIDAndUserID_Return(t *testing.T, returns *types.OrgMember, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByOrgIDAndUserID = func(ctx context.Context, orgID, userID int32) (*types.OrgMember, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
