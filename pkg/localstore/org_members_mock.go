package localstore

import (
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type MockOrgMembers struct {
	GetByOrgIDAndUserID func(ctx context.Context, orgID int32, userID string) (*sourcegraph.OrgMember, error)
}

func (s *MockOrgMembers) MockGetByOrgIDAndUserID_Return(t *testing.T, returns *sourcegraph.OrgMember, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByOrgIDAndUserID = func(ctx context.Context, orgID int32, userID string) (*sourcegraph.OrgMember, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
