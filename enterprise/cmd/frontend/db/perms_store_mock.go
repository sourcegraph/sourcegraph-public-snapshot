package db

import (
	"context"

	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
)

type MockPerms struct {
	Txs                       func(ctx context.Context) (*PermsStore, error)
	LoadUserPermissions       func(ctx context.Context, p *iauthz.UserPermissions) error
	SetRepoPermissions        func(ctx context.Context, p *iauthz.RepoPermissions) error
	SetRepoPendingPermissions func(ctx context.Context, bindIDs []string, p *iauthz.RepoPermissions) error
	GrantPendingPermissionsTx func(ctx context.Context, userID int32, p *iauthz.UserPendingPermissions) error
}
