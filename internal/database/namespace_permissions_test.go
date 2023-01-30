package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCreateNamespacePermission(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.NamespacePermissions()

	user := createUserForNamespacePermission(ctx, t, db, "TestUser")

	np, err := store.Create(ctx, CreateNamespacePermissionOpts{
		Namespace:  "TestNamespace",
		ResourceID: 1,
		UserID:     user.ID,
		Action:     "READ",
	})
	assert.NoError(t, err)
	assert.Equal(t, np.UserID, user.ID)

	// check that the namespace permission created esists
	existingNp, err := store.Get(ctx, GetNamespacePermissionOpts{
		ID: np.ID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, existingNp)
	assert.Equal(t, existingNp.ID, np.ID)
}

func TestDeleteNamespacePermission(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.NamespacePermissions()

	user := createUserForNamespacePermission(ctx, t, db, "user1")

	t.Run("missing ID", func(t *testing.T) {
		err := store.Delete(ctx, DeleteNamespacePermissionOpts{})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing namespace permission id")
	})

	t.Run("existing namespace permissions", func(t *testing.T) {
		np, err := store.Create(ctx, CreateNamespacePermissionOpts{
			Namespace:  "TestNamespace",
			ResourceID: 1,
			UserID:     user.ID,
			Action:     "READ",
		})
		assert.NoError(t, err)

		err = store.Delete(ctx, DeleteNamespacePermissionOpts{
			ID: np.ID,
		})
		assert.NoError(t, err)

		npID := np.ID

		np, err = store.Get(ctx, GetNamespacePermissionOpts{ID: npID})
		assert.Error(t, err)
		assert.Equal(t, err, &NamespacePermissionNotFoundErr{ID: npID})
		assert.Nil(t, np)
	})

	t.Run("non-existent namespace permission", func(t *testing.T) {
		nonExistedNamespacePermissionID := int64(1003)
		err := store.Delete(ctx, DeleteNamespacePermissionOpts{ID: nonExistedNamespacePermissionID})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to delete namespace permission")
	})
}

func TestGetNamespacePermission(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.NamespacePermissions()

	user := createUserForNamespacePermission(ctx, t, db, "user1")

	np, err := store.Create(ctx, CreateNamespacePermissionOpts{
		Namespace:  "TESTNAMESPACE",
		ResourceID: 1,
		UserID:     user.ID,
		Action:     "READ",
	})
	assert.NoError(t, err)

	t.Run("missing ID", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{})
		assert.Nil(t, n)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing namespace permission id")
	})

	t.Run("existing namespace permission", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{ID: np.ID})
		assert.NoError(t, err)
		assert.Equal(t, n.ID, np.ID)
		assert.Equal(t, n.Namespace, np.Namespace)
		assert.Equal(t, n.Action, np.Action)
		assert.Equal(t, n.ResourceID, np.ResourceID)
		assert.Equal(t, n.UserID, np.UserID)
	})

	t.Run("non-existent namespace permission", func(t *testing.T) {
		npID := int64(1003)
		n, err := store.Get(ctx, GetNamespacePermissionOpts{ID: npID})
		assert.Nil(t, n)
		assert.Error(t, err)
		assert.Equal(t, err, &NamespacePermissionNotFoundErr{ID: npID})
	})
}

func createUserForNamespacePermission(ctx context.Context, t *testing.T, db DB, name string) *types.User {
	user, err := db.Users().Create(ctx, NewUser{
		Username: name,
	})
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return user
}
