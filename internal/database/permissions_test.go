package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
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

	totalPerms := 10
	for i := 1; i <= totalPerms; i++ {
		_, err := store.Create(ctx, CreatePermissionOpts{
			Namespace: fmt.Sprintf("PERMISSION-%d", i),
			Action:    "READ",
		})
		assert.NoError(t, err)
	}

	ps, err := store.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, ps, totalPerms)
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
