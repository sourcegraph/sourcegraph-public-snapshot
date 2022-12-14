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

func TestRoleGetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	created, err := store.Create(ctx, "OPERATOR", true)
	if err != nil {
		t.Fatal(err, "unable to create role")
	}

	t.Run("no ID", func(t *testing.T) {
		role, err := store.GetByID(ctx, GetRoleOpts{})
		assert.Error(t, err)
		assert.Nil(t, role)
		assert.Equal(t, err.Error(), "missing id from sql query")
	})

	t.Run("non-existent role", func(t *testing.T) {
		role, err := store.GetByID(ctx, GetRoleOpts{ID: 100})
		assert.Error(t, err)
		assert.EqualError(t, err, "role with ID 100 not found")
		assert.Nil(t, role)
	})

	t.Run("existing role", func(t *testing.T) {
		role, err := store.GetByID(ctx, GetRoleOpts{ID: created.ID})
		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, role.ID, created.ID)
		assert.Equal(t, role.Name, created.Name)
		assert.Equal(t, role.ReadOnly, created.ReadOnly)
		assert.Equal(t, role.CreatedAt, created.CreatedAt)
		assert.Equal(t, role.DeletedAt, created.DeletedAt)
	})
}

func TestRoleList(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	total := createTestRoles(ctx, t, store)

	t.Run("basic no opts", func(t *testing.T) {
		allRoles, err := store.List(ctx, RolesListOptions{})
		assert.NoError(t, err)
		assert.Len(t, allRoles, total)
	})

	t.Run("with pagination", func(t *testing.T) {
		roles, err := store.List(ctx, RolesListOptions{
			LimitOffset: &LimitOffset{Limit: 2, Offset: 1},
		})
		assert.NoError(t, err)
		assert.Len(t, roles, 2)
		assert.Equal(t, roles[0].ID, int32(2))
		assert.Equal(t, roles[1].ID, int32(3))
	})
}

func TestRoleCreate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	_, err := store.Create(ctx, "TESTOLE", true)
	assert.NoError(t, err)
}

func TestRoleCount(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	total := createTestRoles(ctx, t, store)

	count, err := store.Count(ctx, RolesListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, count, total)
}

func TestRoleUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(1234)
		role := types.Role{ID: nonExistentRoleID, Name: "Random Role"}
		updated, err := store.Update(ctx, &role)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Equal(t, err, &RoleNotFoundErr{ID: role.ID})
	})

	t.Run("existing role", func(t *testing.T) {
		role, err := createTestRole(ctx, "TEST ROLE 1", false, t, store)
		assert.NoError(t, err)

		role.Name = "TEST ROLE 2"
		updated, err := store.Update(ctx, role)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, role.Name, "TEST ROLE 2")
	})
}

func TestRoleDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.Roles()

	t.Run("no ID", func(t *testing.T) {
		err := store.Delete(ctx, DeleteRoleOpts{})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing id from sql query")
	})

	t.Run("existing role", func(t *testing.T) {
		role, err := createTestRole(ctx, "TEST ROLE 1", false, t, store)
		assert.NoError(t, err)

		err = store.Delete(ctx, DeleteRoleOpts{role.ID})
		assert.NoError(t, err)

		r, err := store.GetByID(ctx, GetRoleOpts{role.ID})
		assert.Error(t, err)
		assert.Equal(t, err, &RoleNotFoundErr{role.ID})
		assert.Nil(t, r)
	})

	t.Run("non-existent role", func(t *testing.T) {
		nonExistentRoleID := int32(2381)
		err := store.Delete(ctx, DeleteRoleOpts{nonExistentRoleID})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to delete role")
	})
}

func createTestRoles(ctx context.Context, t *testing.T, store RoleStore) int {
	t.Helper()
	totalRoles := 10
	name := "TESTROLE"
	for i := 1; i <= totalRoles; i++ {
		_, err := createTestRole(ctx, fmt.Sprintf("%s-%d", name, i), false, t, store)
		assert.NoError(t, err)
	}
	return totalRoles
}

func createTestRole(ctx context.Context, name string, readonly bool, t *testing.T, store RoleStore) (*types.Role, error) {
	t.Helper()
	return store.Create(ctx, name, readonly)
}
