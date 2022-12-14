package database

import (
	"context"
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
		assert.EqualError(t, err, "role with ID 1 not found")
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
