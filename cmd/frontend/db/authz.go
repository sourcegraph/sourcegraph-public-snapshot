package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type GrantPendingPermissionsArgs struct {
	UserID int32
	BindID string
	Perm   authz.Perms
	Type   authz.PermType
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
