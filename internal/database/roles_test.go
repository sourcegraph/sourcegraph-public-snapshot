package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// func TestRoleList(t *testing.T) {
// 	t.Parallel()
// }

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

func createTestRoles(ctx context.Context, t *testing.T, store RoleStore) int {
	t.Helper()
	totalRoles := 10
	name := "TESTROLE"
	for i := 1; i <= totalRoles; i++ {
		_, err := store.Create(ctx, fmt.Sprintf("%s-%d", name, i), false)
		assert.NoError(t, err)
	}
	return totalRoles
}
