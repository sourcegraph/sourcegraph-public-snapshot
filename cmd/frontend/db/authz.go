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
// or all of the repositories from the candidate list with the given level and type of permissions.
type AuthorizedReposArgs struct {
	// The candidate list of repositories to be verified.
	Repos []*types.Repo
	// The user whose authorization to access the repos is being checked.
	UserID int32
	// The permission level to be verified.
	Perm authz.Perms
	// The type of permissions to be verified.
	Type authz.PermType
	// The type of authz provider to load user permissions.
	Provider authz.ProviderType
}

// AuthzStore contains methods for assigning user permissions.
type AuthzStore interface {
	// GrantPendingPermissions grants pending permissions for a user, it is a no-op in the OSS version.
	GrantPendingPermissions(ctx context.Context, args *GrantPendingPermissionsArgs) error
	// AuthorizedRepos checks if a user is authorized to access repositories in the candidate list.
	// The returned list must be a list of repositories that are authorized to the given user.
	// It is a no-op in the OSS version.
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
