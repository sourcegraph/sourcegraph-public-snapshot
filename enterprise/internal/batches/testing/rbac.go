package testing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func AssignRoleToUser(ctx context.Context, t *testing.T, db database.DB, userID, roleID int32) {
	t.Helper()

	err := db.UserRoles().Assign(ctx, database.AssignUserRoleOpts{
		UserID: userID,
		RoleID: roleID,
	})
	require.NoError(t, err)
}

func CreateTestRole(ctx context.Context, t *testing.T, db database.DB, name string) *types.Role {
	t.Helper()

	role, err := db.Roles().Create(ctx, name, false)
	require.NoError(t, err)

	return role
}

func CreateTestPermission(ctx context.Context, t *testing.T, db database.DB, permission string) *types.Permission {
	t.Helper()

	ns, action, err := rbac.ParsePermissionDisplayName(permission)
	require.NoError(t, err)

	p, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: ns,
		Action:    action,
	})
	require.NoError(t, err)

	return p
}

func AssignPermissionToRole(ctx context.Context, t *testing.T, db database.DB, permissionID, roleID int32) {
	t.Helper()

	err := db.RolePermissions().Assign(ctx, database.AssignRolePermissionOpts{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	require.NoError(t, err)
}
