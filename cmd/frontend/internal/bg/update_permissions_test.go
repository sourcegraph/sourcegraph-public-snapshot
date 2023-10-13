package bg

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdatePermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	allPerms := []*types.Permission{
		{Namespace: rtypes.BatchChangesNamespace, Action: rtypes.BatchChangesReadAction},
		{Namespace: rtypes.BatchChangesNamespace, Action: rtypes.BatchChangesWriteAction},
		{Namespace: rtypes.RepoMetadataNamespace, Action: rtypes.RepoMetadataWriteAction},
		{Namespace: rtypes.OwnershipNamespace, Action: rtypes.OwnershipAssignAction},
	}

	// Updating permissions.
	UpdatePermissions(ctx, logger, db)
	// SITE_ADMINISTRATOR should have all the permissions.
	roleStore := db.Roles()
	adminRole, err := roleStore.Get(ctx, database.GetRoleOpts{Name: string(types.SiteAdministratorSystemRole)})
	require.NoError(t, err)
	permissionStore := db.Permissions()
	adminPermissions, err := permissionStore.List(ctx, database.PermissionListOpts{RoleID: adminRole.ID, PaginationArgs: &database.PaginationArgs{}})
	require.NoError(t, err)
	adminPermissions = clearTimeAndID(adminPermissions)
	assert.ElementsMatch(t, allPerms, adminPermissions)
	// USER should have all the permissions except OWNERSHIP.
	userRole, err := roleStore.Get(ctx, database.GetRoleOpts{Name: string(types.UserSystemRole)})
	require.NoError(t, err)
	userPermissions, err := permissionStore.List(ctx, database.PermissionListOpts{RoleID: userRole.ID, PaginationArgs: &database.PaginationArgs{}})
	require.NoError(t, err)
	userPermissions = clearTimeAndID(userPermissions)
	assert.ElementsMatch(t, allPerms[:3], userPermissions, "unexpected number of permissions")
}

func clearTimeAndID(perms []*types.Permission) []*types.Permission {
	for _, perm := range perms {
		perm.CreatedAt = time.Time{}
		perm.ID = 0
	}
	return perms
}
