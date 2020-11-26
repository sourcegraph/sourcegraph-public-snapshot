package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockAuthz struct {
	GrantPendingPermissions func(ctx context.Context, args *GrantPendingPermissionsArgs) error
	AuthorizedRepos         func(ctx context.Context, args *AuthorizedReposArgs) ([]*types.Repo, error)
	RevokeUserPermissions   func(ctx context.Context, args *RevokeUserPermissionsArgs) error
}
