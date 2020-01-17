package db

import (
	"context"
)

type MockAuthz struct {
	GrantPendingPermissions func(ctx context.Context, args *GrantPendingPermissionsArgs) error
}
