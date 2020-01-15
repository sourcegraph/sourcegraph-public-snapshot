package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
)

type GrantPendingPermissionsArgs struct {
	UserID int32
	BindID string
	Perm   authz.Perms
	Type   authz.PermType
}

// An AuthzStore stores methods for user permissions, they will be set to non-nil in enterprise version.
type AuthzStore interface {
	GrantPendingPermissions(ctx context.Context, args *GrantPendingPermissionsArgs) error
}

// authzStore is a no-op placeholder for the OSS version.
type authzStore struct{}

func (*authzStore) GrantPendingPermissions(_ context.Context, _ *GrantPendingPermissionsArgs) error {
	return nil
}
