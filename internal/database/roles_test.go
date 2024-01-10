package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// The database is already seeded with two roles:
// - USER
// - SITE_ADMINISTRATOR
//
// These roles come by default on any sourcegraph instance and will always exist in the database,
// so we need to account for these roles when accessing the database.
var numberOfSystemRoles = 2

func TestRoleGet(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Roles()

	roleName := "OPERATOR"
	createdRole, err := store.Create(ctx, roleName, true)
	require.NoError(t, err)

	t.Run("without role ID or name", func(t *testing.T) {
		_, err := store.Get(ctx, GetRoleOpts{})
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing id or name")
	})

	t.Run("with role ID", func(t *testing.T) {
		role, err := store.Get(ctx, GetRoleOpts{
			ID: createdRole.ID,
		})
		require.NoError(t, err)
		require.Equal(t, role.ID, createdRole.ID)
		require.Equal(t, role.Name, createdRole.Name)
	})

	t.Run("with role name", func(t *testing.T) {
		role, err := store.Get(ctx, GetRoleOpts{
			Name: roleName,
		})
		require.NoError(t, err)
		require.Equal(t, role.ID, createdRole.ID)
		require.Equal(t, role.Name, createdRole.Name)
	})
}

func TestRoleList(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Roles()

	roles, total := createTestRoles(ctx, t, store)
	user := createTestUserWithoutRoles(t, db, "test-user-1", false)

	err := db.UserRoles().Assign(ctx, AssignUserRoleOpts{
		RoleID: roles[0].ID,
		UserID: user.ID,
	})
	require.NoError(t, err)

	firstParam := 100

	t.Run("all roles", func(t *testing.T) {
		allRoles, err := store.List(ctx, RolesListOptions{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
		})

		require.NoError(t, err)
		require.LessOrEqual(t, len(allRoles), firstParam)
		require.Len(t, allRoles, total+numberOfSystemRoles)
	})

	t.Run("system roles", func(t *testing.T) {
		allSystemRoles, err := store.List(ctx, RolesListOptions{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
			System: true,
		})
		require.NoError(t, err)
		require.Len(t, allSystemRoles, numberOfSystemRoles)
	})

	t.Run("with pagination", func(t *testing.T) {
		firstParam := 2
		roles, err := store.List(ctx, RolesListOptions{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
		})

		require.NoError(t, err)
		require.Len(t, roles, firstParam)
	})

	t.Run("user roles", func(t *testing.T) {
		userRoles, err := store.List(ctx, RolesListOptions{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
			UserID: user.ID,
		})
		require.NoError(t, err)

		require.Len(t, userRoles, 1)
		require.Equal(t, userRoles[0].ID, roles[0].ID)
	})
}

func TestRoleCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Roles()

	_, err := store.Create(ctx, "TESTOLE", true)
	require.NoError(t, err)
}

func TestRoleCount(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Roles()

	user := createTestUserWithoutRoles(t, db, "test-user-1", false)
	roles, total := createTestRoles(ctx, t, store)

	err := db.UserRoles().Assign(ctx, AssignUserRoleOpts{
		RoleID: roles[0].ID,
		UserID: user.ID,
	})
	require.NoError(t, err)

	t.Run("all roles", func(t *testing.T) {
		count, err := store.Count(ctx, RolesListOptions{})

		require.NoError(t, err)
		require.Equal(t, count, total+numberOfSystemRoles)
	})

	t.Run("system roles", func(t *testing.T) {
		count, err := store.Count(ctx, RolesListOptions{
			System: true,
		})

		require.NoError(t, err)
		require.Equal(t, count, numberOfSystemRoles)
	})

	t.Run("user roles", func(t *testing.T) {
		count, err := store.Count(ctx, RolesListOptions{
			UserID: user.ID,
		})

		require.NoError(t, err)
		require.Equal(t, count, 1)
	})
}

func TestRoleUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Roles()

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(1234)
		role := types.Role{ID: nonExistentRoleID, Name: "Random Role"}
		updated, err := store.Update(ctx, &role)
		require.Error(t, err)
		require.Nil(t, updated)
		require.Equal(t, err, &RoleNotFoundErr{ID: role.ID})
	})

	t.Run("existing role", func(t *testing.T) {
		role, err := createTestRole(ctx, "TEST ROLE 1", false, t, store)
		require.NoError(t, err)

		role.Name = "TEST ROLE 2"
		updated, err := store.Update(ctx, role)
		require.NoError(t, err)
		require.NotNil(t, updated)
		require.Equal(t, role.Name, "TEST ROLE 2")
	})
}

func TestRoleDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	store := db.Roles()

	t.Run("no ID", func(t *testing.T) {
		err := store.Delete(ctx, DeleteRoleOpts{})
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing id from sql query")
	})

	t.Run("existing role", func(t *testing.T) {
		role, err := createTestRole(ctx, "TEST ROLE 1", false, t, store)
		require.NoError(t, err)

		err = store.Delete(ctx, DeleteRoleOpts{ID: role.ID})
		require.NoError(t, err)

		r, err := store.Get(ctx, GetRoleOpts{ID: role.ID})
		require.Error(t, err)
		require.Equal(t, err, &RoleNotFoundErr{role.ID})
		require.Nil(t, r)
	})

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(2381)
		err := store.Delete(ctx, DeleteRoleOpts{ID: nonExistentRoleID})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to delete role")
	})
}

func createTestRoles(ctx context.Context, t *testing.T, store RoleStore) ([]*types.Role, int) {
	t.Helper()
	var roles []*types.Role
	totalRoles := 10
	name := "TESTROLE"
	for i := 1; i <= totalRoles; i++ {
		role, err := createTestRole(ctx, fmt.Sprintf("%s-%d", name, i), false, t, store)
		require.NoError(t, err)
		roles = append(roles, role)
	}
	return roles, totalRoles
}

func createTestRole(ctx context.Context, name string, isSystemRole bool, t *testing.T, store RoleStore) (*types.Role, error) {
	t.Helper()
	return store.Create(ctx, name, isSystemRole)
}
