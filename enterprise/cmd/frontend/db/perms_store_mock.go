package db

import (
	"context"

	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
)

type MockPerms struct {
	LoadRepoPermissions        func(ctx context.Context, p *iauthz.RepoPermissions) error
	LoadUserPermissions        func(ctx context.Context, p *iauthz.UserPermissions) error
	LoadUserPendingPermissions func(ctx context.Context, p *iauthz.UserPendingPermissions) error
	SetRepoPermissions         func(ctx context.Context, p *iauthz.RepoPermissions) error
	SetRepoPendingPermissions  func(ctx context.Context, bindIDs []string, p *iauthz.RepoPermissions) error
	ListPendingUsers           func(ctx context.Context) ([]string, error)
}
