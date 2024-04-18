package bg

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUpdatePermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	defaultUserPerms := []*types.Permission{
		{Namespace: rtypes.CodyNamespace, Action: rtypes.CodyAccessAction},
		{Namespace: rtypes.BatchChangesNamespace, Action: rtypes.BatchChangesReadAction},
		{Namespace: rtypes.BatchChangesNamespace, Action: rtypes.BatchChangesWriteAction},
		{Namespace: rtypes.RepoMetadataNamespace, Action: rtypes.RepoMetadataWriteAction},
	}
	defaultAdminPerms := append(defaultUserPerms,
		&types.Permission{Namespace: rtypes.OwnershipNamespace, Action: rtypes.OwnershipAssignAction},
	)

	t.Run("single-tenant", func(t *testing.T) {
		testUpdatePermissionsForCase(t, db, testUpdatePermissionsCase{
			expectedUserPerms:  defaultUserPerms,
			expectedAdminPerms: defaultAdminPerms,
		})
	})

	t.Run("dotcom", func(t *testing.T) {
		dotcom.MockSourcegraphDotComMode(t, true) // helper registers cleanup

		testUpdatePermissionsForCase(t, db, testUpdatePermissionsCase{
			expectedUserPerms: defaultUserPerms,
			expectedAdminPerms: append(defaultAdminPerms,
				// dotcom-only permissions
				&types.Permission{Namespace: rtypes.ProductSubscriptionsNamespace, Action: rtypes.ProductSubscriptionsReadAction},
				&types.Permission{Namespace: rtypes.ProductSubscriptionsNamespace, Action: rtypes.ProductSubscriptionsWriteAction},
			),
		})
	})
}

type testUpdatePermissionsCase struct {
	expectedUserPerms  []*types.Permission
	expectedAdminPerms []*types.Permission
}

// testUpdatePermissionsForCase updates the permissions in the database and
// asserts that the permissions in the database match the expected permissions
// for a testUpdatePermissionsCase.
func testUpdatePermissionsForCase(t *testing.T, db database.DB, tc testUpdatePermissionsCase) {
	t.Helper()

	ctx := context.Background()
	roleStore := db.Roles()
	permissionStore := db.Permissions()

	// Updating permissions.
	UpdatePermissions(ctx, logtest.Scoped(t), db)

	// SITE_ADMINISTRATOR should have all the permissions.
	t.Run("SITE_ADMINISTRATOR", func(t *testing.T) {
		adminRole, err := roleStore.Get(ctx, database.GetRoleOpts{Name: string(types.SiteAdministratorSystemRole)})
		require.NoError(t, err)
		adminPermissions, err := permissionStore.List(ctx, database.PermissionListOpts{RoleID: adminRole.ID, PaginationArgs: &database.PaginationArgs{}})
		require.NoError(t, err)
		assert.ElementsMatch(t, getDisplayNames(tc.expectedAdminPerms), getDisplayNames(adminPermissions))
	})

	// USER should have only the perms designated to users
	t.Run("USER", func(t *testing.T) {
		userRole, err := roleStore.Get(ctx, database.GetRoleOpts{Name: string(types.UserSystemRole)})
		require.NoError(t, err)
		userPermissions, err := permissionStore.List(ctx, database.PermissionListOpts{RoleID: userRole.ID, PaginationArgs: &database.PaginationArgs{}})
		require.NoError(t, err)
		assert.ElementsMatch(t, getDisplayNames(tc.expectedUserPerms), getDisplayNames(userPermissions))
	})
}

func getDisplayNames(perms []*types.Permission) []string {
	var names []string
	for _, perm := range perms {
		names = append(names, perm.DisplayName())
	}
	return names
}
