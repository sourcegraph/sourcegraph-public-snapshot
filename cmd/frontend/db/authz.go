package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
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

// AuthorizedReposArgs contains required arguments to verify if a user is authorized to access some
// or all of repositories from the candidate list with given level and type of permissions.
type AuthorizedReposArgs struct {
	Repos    []*types.Repo      // The candidate list of repositories to be verified.
	UserID   int32              // The user ID that will be used to verify access.
	Perm     authz.Perms        // The permission level to be verified.
	Type     authz.PermType     // The type of permissions to be verified.
	Provider authz.ProviderType // The type of authz provider to load user permissions.
}

// AuthzStore contains methods for assigning user permissions.
type AuthzStore interface {
	// GrantPendingPermissions grants pending permissions for a user, it is a no-op in the OSS version.
	GrantPendingPermissions(ctx context.Context, args *GrantPendingPermissionsArgs) error
	AuthorizedRepos(ctx context.Context, args *AuthorizedReposArgs) ([]*types.Repo, error)
}

// authzStore is a no-op placeholder for the OSS version.
type authzStore struct{}

func (*authzStore) GrantPendingPermissions(_ context.Context, _ *GrantPendingPermissionsArgs) error {
	return nil
}

func (*authzStore) AuthorizedRepos(_ context.Context, _ *AuthorizedReposArgs) ([]*types.Repo, error) {
	return []*types.Repo{}, nil
}
