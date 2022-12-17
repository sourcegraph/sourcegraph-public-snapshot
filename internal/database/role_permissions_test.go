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

func TestRolePermissionCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	r, p := createRoleAndPermission(ctx, t, db)

	t.Run("without permission id", func(t *testing.T) {
		rp, err := store.Create(ctx, CreateRolePermissionOpts{
			RoleID: r.ID,
		})
		assert.Nil(t, rp)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing permission id")
	})

	t.Run("without role id", func(t *testing.T) {
		rp, err := store.Create(ctx, CreateRolePermissionOpts{
			PermissionID: p.ID,
		})
		assert.Nil(t, rp)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing role id")
	})

	t.Run("with correct args", func(t *testing.T) {
		rp, err := store.Create(ctx, CreateRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, rp)
		assert.Equal(t, rp.RoleID, r.ID)
		assert.Equal(t, rp.PermissionID, p.ID)
	})
}

func TestRolePermissionGetByRoleIDAndPermissionID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	r, p := createRoleAndPermission(ctx, t, db)
	_, err := store.Create(ctx, CreateRolePermissionOpts{
		RoleID:       r.ID,
		PermissionID: p.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("without permission ID", func(t *testing.T) {
		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			RoleID: r.ID,
		})
		assert.Nil(t, rp)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing permission id")
	})

	t.Run("without role ID", func(t *testing.T) {
		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: p.ID,
		})
		assert.Nil(t, rp)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing role id")
	})

	t.Run("non existent role id and permission id", func(t *testing.T) {
		pid := int32(1083)
		rid := int32(2342)
		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: pid,
			RoleID:       rid,
		})

		assert.Nil(t, rp)
		assert.Error(t, err)
		assert.Equal(t, err, &RolePermissionNotFoundErr{PermissionID: pid, RoleID: rid})
	})

	t.Run("with correct args", func(t *testing.T) {
		rp, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: p.ID,
			RoleID:       r.ID,
		})

		assert.NoError(t, err)
		assert.Equal(t, rp.RoleID, r.ID)
		assert.Equal(t, rp.PermissionID, p.ID)
	})
}

func TestRolePermissionGetByRoleID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	r := createTestRoleForRolePermission(ctx, "TEST ROLE", t, db)

	totalRolePermissions := 5
	for i := 1; i <= totalRolePermissions; i++ {
		p := createTestPermissionForRolePermission(ctx, "BATCH CHANGES", fmt.Sprintf("action-%d", i), t, db)
		_, err := store.Create(ctx, CreateRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})

		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("without role ID", func(t *testing.T) {
		rp, err := store.GetByRoleID(ctx, GetRolePermissionOpts{})

		assert.Nil(t, rp)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing role id")
	})

	t.Run("with correct args", func(t *testing.T) {
		rps, err := store.GetByRoleID(ctx, GetRolePermissionOpts{
			RoleID: r.ID,
		})

		assert.NoError(t, err)
		assert.Len(t, rps, totalRolePermissions)

		for _, rp := range rps {
			assert.Equal(t, rp.RoleID, r.ID)
		}
	})
}

func TestRolePermissionGetByPermissionID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	p := createTestPermissionForRolePermission(ctx, "BATCH CHANGES", "READ", t, db)

	totalRolePermissions := 5
	for i := 1; i <= totalRolePermissions; i++ {
		r := createTestRoleForRolePermission(ctx, fmt.Sprintf("TEST ROLE-%d", i), t, db)
		_, err := store.Create(ctx, CreateRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})

		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("without permission ID", func(t *testing.T) {
		rp, err := store.GetByPermissionID(ctx, GetRolePermissionOpts{})

		assert.Nil(t, rp)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing permission id")
	})

	t.Run("with correct args", func(t *testing.T) {
		rps, err := store.GetByPermissionID(ctx, GetRolePermissionOpts{
			PermissionID: p.ID,
		})

		assert.NoError(t, err)
		assert.Len(t, rps, totalRolePermissions)

		for _, rp := range rps {
			assert.Equal(t, rp.PermissionID, p.ID)
		}
	})
}

func TestRolePermissionDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.RolePermissions()

	r, p := createRoleAndPermission(ctx, t, db)

	_, err := store.Create(ctx, CreateRolePermissionOpts{
		RoleID:       r.ID,
		PermissionID: p.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("missing permission id", func(t *testing.T) {
		err := store.Delete(ctx, DeleteRolePermissionOpts{})

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing permission id")
	})

	t.Run("missing role id", func(t *testing.T) {
		err := store.Delete(ctx, DeleteRolePermissionOpts{
			PermissionID: p.ID,
		})

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing role id")
	})

	t.Run("with non-existent role permission", func(t *testing.T) {
		roleID := int32(1234)
		permissionID := int32(4321)

		err := store.Delete(ctx, DeleteRolePermissionOpts{
			RoleID:       roleID,
			PermissionID: permissionID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to delete role permission")
	})

	t.Run("with existing role permission", func(t *testing.T) {
		err := store.Delete(ctx, DeleteRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
		assert.NoError(t, err)

		ur, err := store.GetByRoleIDAndPermissionID(ctx, GetRolePermissionOpts{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
		assert.Nil(t, ur)
		assert.Error(t, err)
		assert.Equal(t, err, &RolePermissionNotFoundErr{
			RoleID:       r.ID,
			PermissionID: p.ID,
		})
	})
}

func createTestPermissionForRolePermission(ctx context.Context, namespace, action string, t *testing.T, db DB) *types.Permission {
	t.Helper()
	p, err := db.Permissions().Create(ctx, CreatePermissionOpts{
		Namespace: namespace,
		Action:    action,
	})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func createRoleAndPermission(ctx context.Context, t *testing.T, db DB) (*types.Role, *types.Permission) {
	t.Helper()
	permission := createTestPermissionForRolePermission(ctx, "BATCHCHANGE", "READ", t, db)
	role := createTestRoleForRolePermission(ctx, "TEST ROLE", t, db)
	return role, permission
}

func createTestRoleForRolePermission(ctx context.Context, name string, t *testing.T, db DB) *types.Role {
	t.Helper()
	r, err := db.Roles().Create(ctx, name, false)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
