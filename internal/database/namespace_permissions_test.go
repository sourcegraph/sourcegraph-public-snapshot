package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCreateNamespacePermission(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(t))
	store := db.NamespacePermissions()

	user := createUserForNamespacePermission(ctx, t, db, "TestUser")

	t.Run("missing resource id", func(t *testing.T) {
		np, err := store.Create(ctx, CreateNamespacePermissionOpts{
			Namespace: rtypes.BatchChangesNamespace,
			UserID:    user.ID,
		})
		require.Nil(t, np)
		require.Error(t, err)
		require.ErrorContains(t, err, "resource id is required")
	})

	t.Run("missing user id", func(t *testing.T) {
		np, err := store.Create(ctx, CreateNamespacePermissionOpts{
			Namespace:  rtypes.BatchChangesNamespace,
			ResourceID: 1,
		})
		require.Nil(t, np)
		require.Error(t, err)
		require.ErrorContains(t, err, "user id is required")
	})

	t.Run("missing namespace", func(t *testing.T) {
		np, err := store.Create(ctx, CreateNamespacePermissionOpts{
			ResourceID: 1,
			UserID:     user.ID,
		})

		require.Nil(t, np)
		require.Error(t, err)
		require.ErrorContains(t, err, "valid namespace is required")
	})

	t.Run("invalid namespace", func(t *testing.T) {
		np, err := store.Create(ctx, CreateNamespacePermissionOpts{
			Namespace:  rtypes.PermissionNamespace("TEST_NAMESPACE"),
			ResourceID: 1,
			UserID:     user.ID,
		})

		require.Nil(t, np)
		require.Error(t, err)
		require.ErrorContains(t, err, "valid namespace is required")
	})

	t.Run("success", func(t *testing.T) {
		np, err := store.Create(ctx, CreateNamespacePermissionOpts{
			Namespace:  rtypes.BatchChangesNamespace,
			ResourceID: 1,
			UserID:     user.ID,
		})
		require.NoError(t, err)
		require.Equal(t, np.UserID, user.ID)

		// check that the namespace permission created esists
		existingNp, err := store.Get(ctx, GetNamespacePermissionOpts{
			ID: np.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, existingNp)
		require.Equal(t, existingNp.ID, np.ID)
	})
}

func TestDeleteNamespacePermission(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(t))
	store := db.NamespacePermissions()

	user := createUserForNamespacePermission(ctx, t, db, "user1")

	t.Run("missing ID", func(t *testing.T) {
		err := store.Delete(ctx, DeleteNamespacePermissionOpts{})
		require.Error(t, err)
		require.ErrorContains(t, err, "namespace permission id is required")
	})

	t.Run("existing namespace permissions", func(t *testing.T) {
		np, err := store.Create(ctx, CreateNamespacePermissionOpts{
			Namespace:  rtypes.BatchChangesNamespace,
			ResourceID: 1,
			UserID:     user.ID,
		})
		require.NoError(t, err)

		err = store.Delete(ctx, DeleteNamespacePermissionOpts{
			ID: np.ID,
		})
		require.NoError(t, err)

		npID := np.ID

		np, err = store.Get(ctx, GetNamespacePermissionOpts{ID: npID})
		require.Error(t, err)
		require.Equal(t, err, &NamespacePermissionNotFoundErr{ID: npID})
		require.Nil(t, np)
	})

	t.Run("non-existent namespace permission", func(t *testing.T) {
		nonExistedNamespacePermissionID := int64(1003)
		err := store.Delete(ctx, DeleteNamespacePermissionOpts{ID: nonExistedNamespacePermissionID})

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to delete namespace permission")
	})
}

func TestGetNamespacePermission(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(t))
	store := db.NamespacePermissions()

	user := createUserForNamespacePermission(ctx, t, db, "user1")

	np, err := store.Create(ctx, CreateNamespacePermissionOpts{
		Namespace:  rtypes.BatchChangesNamespace,
		ResourceID: 1,
		UserID:     user.ID,
	})
	require.NoError(t, err)

	t.Run("missing namespace permission ID", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing namespace permission query")
	})

	t.Run("missing namespace", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{
			UserID:     user.ID,
			ResourceID: 1,
		})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing namespace permission query")
	})

	t.Run("invalid namespace", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{
			UserID:     user.ID,
			ResourceID: 1,
			Namespace:  rtypes.PermissionNamespace("TEST-NAMESPACE"),
		})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing namespace permission query")
	})

	t.Run("missing resource id", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{
			Namespace: rtypes.BatchChangesNamespace,
			UserID:    user.ID,
		})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing namespace permission query")
	})

	t.Run("missing user id", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{
			Namespace:  rtypes.BatchChangesNamespace,
			ResourceID: 1,
		})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equal(t, err.Error(), "missing namespace permission query")
	})

	t.Run("existing namespace permission (via ID)", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{ID: np.ID})

		require.NoError(t, err)
		require.Equal(t, n.ID, np.ID)
		require.Equal(t, n.Namespace, np.Namespace)
		require.Equal(t, n.ResourceID, np.ResourceID)
		require.Equal(t, n.UserID, np.UserID)
	})

	t.Run("existing namespace permission", func(t *testing.T) {
		n, err := store.Get(ctx, GetNamespacePermissionOpts{
			Namespace:  np.Namespace,
			ResourceID: np.ResourceID,
			UserID:     np.UserID,
		})

		require.NoError(t, err)
		require.Equal(t, n.ID, np.ID)
		require.Equal(t, n.Namespace, np.Namespace)
		require.Equal(t, n.ResourceID, np.ResourceID)
		require.Equal(t, n.UserID, np.UserID)
	})

	t.Run("non-existent namespace permission", func(t *testing.T) {
		npID := int64(1003)
		n, err := store.Get(ctx, GetNamespacePermissionOpts{ID: npID})

		require.Nil(t, n)
		require.Error(t, err)
		require.Equal(t, err, &NamespacePermissionNotFoundErr{ID: npID})
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
