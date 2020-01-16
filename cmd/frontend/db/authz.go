package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// GrantPendingPermissionsArgs contains required arguments to grant pending permissions for a user.
// Both "Username" and "Email" could be supplied but only one of it will be used according to the
// site configuration.
// 🚨 SECURITY: It is the caller's responsibility to ensure the supplied email is verified.
type GrantPendingPermissionsArgs struct {
	UserID   int32
	Username string
	Email    string
	Perm     authz.Perms
	Type     authz.PermType
}

type AuthorizedReposArgs struct {
	Repos    []*types.Repo
	UserID   int32
	Perm     authz.Perms
	Type     authz.PermType
	Provider authz.ProviderType
}

// An AuthzStore stores methods for user permissions, they will be no-op in OSS version.
type AuthzStore interface {
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
