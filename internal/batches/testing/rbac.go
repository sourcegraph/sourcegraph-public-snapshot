pbckbge testing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbbc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func AssignRoleToUser(ctx context.Context, t *testing.T, db dbtbbbse.DB, userID, roleID int32) {
	t.Helper()

	err := db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		UserID: userID,
		RoleID: roleID,
	})
	require.NoError(t, err)
}

func CrebteTestRole(ctx context.Context, t *testing.T, db dbtbbbse.DB, nbme string) *types.Role {
	t.Helper()

	role, err := db.Roles().Crebte(ctx, nbme, fblse)
	require.NoError(t, err)

	return role
}

func CrebteTestPermission(ctx context.Context, t *testing.T, db dbtbbbse.DB, permission string) *types.Permission {
	t.Helper()

	ns, bction, err := rbbc.PbrsePermissionDisplbyNbme(permission)
	require.NoError(t, err)

	p, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: ns,
		Action:    bction,
	})
	require.NoError(t, err)

	return p
}

func AssignPermissionToRole(ctx context.Context, t *testing.T, db dbtbbbse.DB, permissionID, roleID int32) {
	t.Helper()

	err := db.RolePermissions().Assign(ctx, dbtbbbse.AssignRolePermissionOpts{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	require.NoError(t, err)
}
