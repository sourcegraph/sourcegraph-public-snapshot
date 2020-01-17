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
	// The user ID that will be used to bind pending permissions.
	UserID int32
	// The username that will be used as bind ID.
	Username string
	// The verified email address that will be used as bind ID.
	// 🚨 SECURITY: It is the caller's responsibility to ensure the email is verified.
	VerifiedEmail string
	// The permission level to be granted.
	Perm authz.Perms
	// The type of permissions to be granted.
	Type authz.PermType
}

// AuthzStore contains methods for assigning user permissions.
type AuthzStore interface {
	// GrantPendingPermissions grants pending permissions for a user, it is a no-op in the OSS version.
	GrantPendingPermissions(ctx context.Context, args *GrantPendingPermissionsArgs) error
}

// authzStore is a no-op placeholder for the OSS version.
type authzStore struct{}

func (*authzStore) GrantPendingPermissions(_ context.Context, _ *GrantPendingPermissionsArgs) error {
	return nil
}
