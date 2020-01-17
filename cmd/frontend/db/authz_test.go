package db

import "context"

var _ AuthzStore = &mockAuthzStore{}

// mockAuthzStore is a mock struct that implements AuthzStore interface and only calls mock methods
// when they are not nil.
type mockAuthzStore struct{}

func (*mockAuthzStore) GrantPendingPermissions(ctx context.Context, args *GrantPendingPermissionsArgs) error {
	if Mocks.Authz.GrantPendingPermissions == nil {
		return nil
	}
	return Mocks.Authz.GrantPendingPermissions(ctx, args)
}
