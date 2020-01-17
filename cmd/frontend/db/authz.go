package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
)

// GrantPendingPermissionsArgs contains required arguments to grant pending permissions for a user.
// Both "Username" and "Email" could be supplied but only one of them will be used according to the
// site configuration.
// 🚨 SECURITY: It is the caller's responsibility to ensure the supplied email is verified.
type GrantPendingPermissionsArgs struct {
	UserID   int32          // The user ID that will be used to bind pending permissions.
	Username string         // The username that will be used as bind ID.
	Email    string         // The email address that will be used as bind ID. It is the caller's responsibility to ensure the email is verified.
	Perm     authz.Perms    // The permission level to be granted.
	Type     authz.PermType // The type of permissions to be granted.
}

// AuthzStore contains methods for assigning user permissions.
type AuthzStore interface {
	GrantPendingPermissions(ctx context.Context, args *GrantPendingPermissionsArgs) error
}

// authzStore is a no-op placeholder for the OSS version.
type authzStore struct{}

func (*authzStore) GrantPendingPermissions(_ context.Context, _ *GrantPendingPermissionsArgs) error {
	return nil
}
