package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

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
		Namespace: "BATCHCHANGES",
		Action:    "READ",
	})
	if err != nil {
		t.Fatal(err, "unable to create permission")
	}

	t.Run("no ID", func(t *testing.T) {
		p, err := store.GetByID(ctx, GetPermissionOpts{})
		assert.Error(t, err)
		assert.Nil(t, p)
		assert.Equal(t, err.Error(), "missing id from sql query")
	})

	t.Run("non-existent permission", func(t *testing.T) {
		p, err := store.GetByID(ctx, GetPermissionOpts{ID: 100})
		assert.Error(t, err)
		assert.EqualError(t, err, "permission with ID 100 not found")
		assert.Nil(t, p)
	})

	t.Run("existing permission", func(t *testing.T) {
		permission, err := store.GetByID(ctx, GetPermissionOpts{ID: created.ID})
		assert.NoError(t, err)
		assert.NotNil(t, permission)
		assert.Equal(t, permission.ID, created.ID)
		assert.Equal(t, permission.Namespace, created.Namespace)
		assert.Equal(t, permission.Action, created.Action)
	})
}

func TestPermissionCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	_, err := store.Create(ctx, CreatePermissionOpts{
		Namespace: "BATCHCHANGES",
		Action:    "READ",
	})
	assert.NoError(t, err)
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

		assert.NoError(t, err)
		assert.Len(t, ps, totalPerms)
		assert.LessOrEqual(t, len(ps), firstParam)
	})

	t.Run("with pagination", func(t *testing.T) {
		firstParam := 2
		ps, err := store.List(ctx, PermissionListOpts{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
		})

		assert.NoError(t, err)
		assert.Len(t, ps, firstParam)
	})

	t.Run("role association", func(t *testing.T) {
		ps, err := store.List(ctx, PermissionListOpts{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
			RoleID: role.ID,
		})

		assert.NoError(t, err)
		assert.Len(t, ps, 2)
	})

	t.Run("user association", func(t *testing.T) {
		ps, err := store.List(ctx, PermissionListOpts{
			PaginationArgs: &PaginationArgs{
				First: &firstParam,
			},
			UserID: user.ID,
		})

		assert.NoError(t, err)
		assert.Len(t, ps, 2)
	})
}

func TestPermissionDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	p, err := store.Create(ctx, CreatePermissionOpts{
		Namespace: "BATCHCHANGES",
		Action:    "READ",
	})
	assert.NoError(t, err)

	t.Run("no ID", func(t *testing.T) {
		err := store.Delete(ctx, DeletePermissionOpts{})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing id from sql query")
	})

	t.Run("existing role", func(t *testing.T) {
		err = store.Delete(ctx, DeletePermissionOpts{p.ID})
		assert.NoError(t, err)

		deleted, err := store.GetByID(ctx, GetPermissionOpts{ID: p.ID})
		assert.Nil(t, deleted)
		assert.Error(t, err)
		assert.Equal(t, err, &PermissionNotFoundErr{p.ID})
	})

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(2381)
		err := store.Delete(ctx, DeletePermissionOpts{nonExistentRoleID})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to delete permission")
	})
}

func TestPermissionBulkCreate(t *testing.T) {
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
			Action:    action,
			Namespace: fmt.Sprintf("namespace-%d", i),
		})
	}

	ps, err := store.BulkCreate(ctx, perms)
	assert.NoError(t, err)
	assert.NotNil(t, ps)
	assert.Len(t, ps, 5)
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
			Action:    action,
			Namespace: fmt.Sprintf("namespace-for-deletion-%d", i),
		})
	}

	ps, err := store.BulkCreate(ctx, perms)
	assert.NoError(t, err)

	var permsToBeDeleted []DeletePermissionOpts
	for _, p := range ps {
		permsToBeDeleted = append(permsToBeDeleted, DeletePermissionOpts{
			ID: p.ID,
		})
	}

	t.Run("no options provided", func(t *testing.T) {
		err = store.BulkDelete(ctx, []DeletePermissionOpts{})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing ids from sql query")
	})

	t.Run("non existent roles", func(t *testing.T) {
		err = store.BulkDelete(ctx, []DeletePermissionOpts{
			{ID: 109},
		})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "failed to delete permissions")
	})

	t.Run("existing roles", func(t *testing.T) {
		err = store.BulkDelete(ctx, permsToBeDeleted)
		assert.NoError(t, err)

		// check if the first permission exists in the database
		deleted, err := store.GetByID(ctx, GetPermissionOpts{ID: ps[0].ID})
		assert.Nil(t, deleted)
		assert.Error(t, err)
		assert.Equal(t, err, &PermissionNotFoundErr{ps[0].ID})
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

		assert.NoError(t, err)
		assert.Equal(t, count, totalPerms)
	})

	t.Run("role permissions", func(t *testing.T) {
		count, err := store.Count(ctx, PermissionListOpts{
			RoleID: role.ID,
		})

		assert.NoError(t, err)
		assert.Equal(t, count, 2)
	})

	t.Run("user permissions", func(t *testing.T) {
		count, err := store.Count(ctx, PermissionListOpts{
			UserID: user.ID,
		})

		assert.NoError(t, err)
		assert.Equal(t, count, 2)
	})
}

func TestPermissionFetchAll(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Permissions()

	_, _, totalPerms := seedPermissionDataForList(ctx, t, store, db)

	perms, err := store.FetchAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, perms, totalPerms)
}

func seedPermissionDataForList(ctx context.Context, t *testing.T, store PermissionStore, db DB) (*types.Role, *types.User, int) {
	t.Helper()

	perms, totalPerms := createTestPermissions(ctx, t, store)
	user := createTestUserForUserRole(ctx, "test@test.com", "test-user-1", t, db)
	role, err := createTestRole(ctx, "TEST-ROLE", false, t, db.Roles())
	assert.NoError(t, err)

	_, err = db.RolePermissions().Create(ctx, CreateRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perms[0].ID,
	})
	assert.NoError(t, err)

	_, err = db.RolePermissions().Create(ctx, CreateRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perms[1].ID,
	})
	assert.NoError(t, err)

	_, err = db.UserRoles().Create(ctx, CreateUserRoleOpts{
		RoleID: role.ID,
		UserID: user.ID,
	})
	assert.NoError(t, err)

	return role, user, totalPerms
}

func createTestPermissions(ctx context.Context, t *testing.T, store PermissionStore) ([]*types.Permission, int) {
	t.Helper()

	var permissions []*types.Permission

	totalPerms := 10
	for i := 1; i <= totalPerms; i++ {
		permission, err := store.Create(ctx, CreatePermissionOpts{
			Namespace: fmt.Sprintf("PERMISSION-%d", i),
			Action:    "READ",
		})
		assert.NoError(t, err)
		permissions = append(permissions, permission)
	}

	return permissions, totalPerms
}
