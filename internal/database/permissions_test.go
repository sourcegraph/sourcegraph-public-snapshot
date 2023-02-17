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

func TestPermissionGetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	created, err := store.Create(ctx, CreatePermissionOpts{
		Namespace: types.BatchChangesNamespace,
		Action:    "READ",
	})
	if err != nil {
		t.Fatal(err, "unable to create permission")
	}

	t.Run("no ID", func(t *testing.T) {
		p, err := store.GetByID(ctx, GetPermissionOpts{})
		require.Error(t, err)
		require.Nil(t, p)
		require.Equal(t, err.Error(), "missing id from sql query")
	})

	t.Run("non-existent permission", func(t *testing.T) {
		p, err := store.GetByID(ctx, GetPermissionOpts{ID: 100})
		require.Error(t, err)
		require.EqualError(t, err, "permission with ID 100 not found")
		require.Nil(t, p)
	})

	t.Run("existing permission", func(t *testing.T) {
		permission, err := store.GetByID(ctx, GetPermissionOpts{ID: created.ID})
		require.NoError(t, err)
		require.NotNil(t, permission)
		require.Equal(t, permission.ID, created.ID)
		require.Equal(t, permission.Namespace, created.Namespace)
		require.Equal(t, permission.Action, created.Action)
	})
}

func TestPermissionCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	t.Run("invalid namespace", func(t *testing.T) {
		p, err := store.Create(ctx, CreatePermissionOpts{
			Namespace: types.PermissionNamespace("TEST-NAMESPACE"),
			Action:    "READ",
		})

		require.Nil(t, p)
		require.Error(t, err)
		require.ErrorContains(t, err, "valid action and namespace is required")
	})

	t.Run("missing namespace", func(t *testing.T) {
		p, err := store.Create(ctx, CreatePermissionOpts{
			Action: "READ",
		})

		require.Nil(t, p)
		require.Error(t, err)
		require.ErrorContains(t, err, "valid action and namespace is required")
	})

	t.Run("missing action", func(t *testing.T) {
		p, err := store.Create(ctx, CreatePermissionOpts{
			Namespace: types.PermissionNamespace("TEST-NAMESPACE"),
		})

		require.Nil(t, p)
		require.Error(t, err)
		require.ErrorContains(t, err, "valid action and namespace is required")
	})

	t.Run("success", func(t *testing.T) {
		p, err := store.Create(ctx, CreatePermissionOpts{
			Namespace: types.BatchChangesNamespace,
			Action:    "READ",
		})

		require.NotNil(t, p)
		require.NoError(t, err)
	})
}

func TestPermissionList(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	role, user, totalPerms := seedPermissionDataForList(ctx, t, store, db)
	firstParam := 100

	t.Run("all permissions", func(t *testing.T) {
		ps, err := store.List(ctx, PermissionListOpts{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
		})

		require.NoError(t, err)
		require.Len(t, ps, totalPerms)
		require.LessOrEqual(t, len(ps), firstParam)
	})

	t.Run("with pagination", func(t *testing.T) {
		firstParam := 2
		ps, err := store.List(ctx, PermissionListOpts{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
		})

		require.NoError(t, err)
		require.Len(t, ps, firstParam)
	})

	t.Run("role association", func(t *testing.T) {
		ps, err := store.List(ctx, PermissionListOpts{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
			RoleID: role.ID,
		})

		require.NoError(t, err)
		require.Len(t, ps, 2)
	})

	t.Run("user association", func(t *testing.T) {
		ps, err := store.List(ctx, PermissionListOpts{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
			UserID: user.ID,
		})

		require.NoError(t, err)
		require.Len(t, ps, 2)
	})
}

func TestPermissionDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	p, err := store.Create(ctx, CreatePermissionOpts{
		Namespace: types.BatchChangesNamespace,
		Action:    "READ",
	})
	require.NoError(t, err)

	t.Run("no ID", func(t *testing.T) {
		err := store.Delete(ctx, DeletePermissionOpts{})
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing id from sql query")
	})

	t.Run("existing role", func(t *testing.T) {
		err = store.Delete(ctx, DeletePermissionOpts{p.ID})
		require.NoError(t, err)

		deleted, err := store.GetByID(ctx, GetPermissionOpts{ID: p.ID})
		require.Nil(t, deleted)
		require.Error(t, err)
		require.Equal(t, err, &PermissionNotFoundErr{p.ID})
	})

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(2381)
		err := store.Delete(ctx, DeletePermissionOpts{nonExistentRoleID})
		require.Error(t, err)
		require.ErrorContains(t, err, "failed to delete permission")
	})
}

func TestPermissionBulkCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	t.Run("invalid namespace", func(t *testing.T) {
		opts := []CreatePermissionOpts{
			{Action: "READ", Namespace: types.PermissionNamespace("TEST-NAMESPACE")},
		}

		ps, err := store.BulkCreate(ctx, opts)
		require.ErrorContains(t, err, "valid namespace is required")
		require.Nil(t, ps)
	})

	t.Run("success", func(t *testing.T) {
		noOfPerms := 5
		var opts []CreatePermissionOpts
		for i := 1; i <= noOfPerms; i++ {
			var action string
			if i%2 == 0 {
				action = "READ"
			} else {
				action = "WRITE"
			}
			opts = append(opts, CreatePermissionOpts{
				Action:    fmt.Sprintf("%s-%d", action, i),
				Namespace: types.BatchChangesNamespace,
			})
		}

		ps, err := store.BulkCreate(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, ps)
		require.Len(t, ps, noOfPerms)
	})
}

func TestPermissionBulkDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	var perms []CreatePermissionOpts
	for i := 1; i <= 5; i++ {
		var action string
		if i%2 == 0 {
			action = "READ"
		} else {
			action = "WRITE"
		}
		perms = append(perms, CreatePermissionOpts{
			Action:    fmt.Sprintf("%s-%d", action, i),
			Namespace: types.BatchChangesNamespace,
		})
	}

	ps, err := store.BulkCreate(ctx, perms)
	require.NoError(t, err)

	var permsToBeDeleted []DeletePermissionOpts
	for _, p := range ps {
		permsToBeDeleted = append(permsToBeDeleted, DeletePermissionOpts{
			ID: p.ID,
		})
	}

	t.Run("no options provided", func(t *testing.T) {
		err = store.BulkDelete(ctx, []DeletePermissionOpts{})
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing ids from sql query")
	})

	t.Run("non existent roles", func(t *testing.T) {
		err = store.BulkDelete(ctx, []DeletePermissionOpts{
			{ID: 109},
		})
		require.Error(t, err)
		require.Equal(t, err.Error(), "failed to delete permissions")
	})

	t.Run("existing roles", func(t *testing.T) {
		err = store.BulkDelete(ctx, permsToBeDeleted)
		require.NoError(t, err)

		// check if the first permission exists in the database
		deleted, err := store.GetByID(ctx, GetPermissionOpts{ID: ps[0].ID})
		require.Nil(t, deleted)
		require.Error(t, err)
		require.Equal(t, err, &PermissionNotFoundErr{ps[0].ID})
	})
}

func TestPermissionCount(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	role, user, totalPerms := seedPermissionDataForList(ctx, t, store, db)

	t.Run("all permissions", func(t *testing.T) {
		count, err := store.Count(ctx, PermissionListOpts{})

		require.NoError(t, err)
		require.Equal(t, count, totalPerms)
	})

	t.Run("role permissions", func(t *testing.T) {
		count, err := store.Count(ctx, PermissionListOpts{
			RoleID: role.ID,
		})

		require.NoError(t, err)
		require.Equal(t, count, 2)
	})

	t.Run("user permissions", func(t *testing.T) {
		count, err := store.Count(ctx, PermissionListOpts{
			UserID: user.ID,
		})

		require.NoError(t, err)
		require.Equal(t, count, 2)
	})
}

func TestPermissionFetchAll(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	_, _, totalPerms := seedPermissionDataForList(ctx, t, store, db)

	perms, err := store.FetchAll(ctx)

	require.NoError(t, err)
	require.Len(t, perms, totalPerms)
}

func seedPermissionDataForList(ctx context.Context, t *testing.T, store PermissionStore, db DB) (*types.Role, *types.User, int) {
	t.Helper()

	perms, totalPerms := createTestPermissions(ctx, t, store)
	user := createTestUserWithoutRoles(t, db, "test-user-1", false)
	role, err := createTestRole(ctx, "TEST-ROLE", false, t, db.Roles())
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perms[0].ID,
	})
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perms[1].ID,
	})
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: user.ID,
	})
	require.NoError(t, err)

	return role, user, totalPerms
}

func createTestPermissions(ctx context.Context, t *testing.T, store PermissionStore) ([]*types.Permission, int) {
	t.Helper()

	var permissions []*types.Permission

	totalPerms := 10
	for i := 1; i <= totalPerms; i++ {
		permission, err := store.Create(ctx, CreatePermissionOpts{
			Namespace: types.BatchChangesNamespace,
			Action:    fmt.Sprintf("READ-%d", i),
		})
		require.NoError(t, err)
		permissions = append(permissions, permission)
	}

	return permissions, totalPerms
}
