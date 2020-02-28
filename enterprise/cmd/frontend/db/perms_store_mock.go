package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
)

type MockPerms struct {
	LoadRepoPermissions        func(ctx context.Context, p *authz.RepoPermissions) error
	LoadUserPermissions        func(ctx context.Context, p *authz.UserPermissions) error
	LoadUserPendingPermissions func(ctx context.Context, p *authz.UserPendingPermissions) error
	SetRepoPermissions         func(ctx context.Context, p *authz.RepoPermissions) error
	SetRepoPendingPermissions  func(ctx context.Context, bindIDs []string, p *authz.RepoPermissions) error
	ListPendingUsers           func(ctx context.Context) ([]string, error)
}
