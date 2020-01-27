package db

import (
	"context"

	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
)

type MockPerms struct {
	Transact                func(ctx context.Context) (*PermsStore, error)
	LoadUserPermissions     func(ctx context.Context, p *iauthz.UserPermissions) error
	GrantPendingPermissions func(ctx context.Context, userID int32, p *iauthz.UserPendingPermissions) error
}
