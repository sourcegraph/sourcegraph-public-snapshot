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

func TestUserRoleCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, role := createUserAndRole(ctx, t, db)

	t.Run("without user id", func(t *testing.T) {
		ur, err := store.Create(ctx, CreateUserRoleOpts{
			RoleID: role.ID,
		})
		assert.Nil(t, ur)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing user id")
	})

	t.Run("without role id", func(t *testing.T) {
		ur, err := store.Create(ctx, CreateUserRoleOpts{
			UserID: user.ID,
		})
		assert.Nil(t, ur)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing role id")
	})

	t.Run("with correct args", func(t *testing.T) {
		ur, err := store.Create(ctx, CreateUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, ur)
		assert.Equal(t, ur.RoleID, role.ID)
		assert.Equal(t, ur.UserID, user.ID)
	})
}

func TestUserRoleDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, role := createUserAndRole(ctx, t, db)

	// create a user role
	_, err := store.Create(ctx, CreateUserRoleOpts{
		RoleID: role.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("missing user id", func(t *testing.T) {
		err := store.Delete(ctx, DeleteUserRoleOpts{
			RoleID: role.ID,
		})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing user id")
	})

	t.Run("missing role id", func(t *testing.T) {
		err := store.Delete(ctx, DeleteUserRoleOpts{
			UserID: user.ID,
		})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing role id")
	})

	t.Run("with existing user role", func(t *testing.T) {
		err := store.Delete(ctx, DeleteUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		assert.NoError(t, err)

		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		assert.Nil(t, ur)
		assert.Error(t, err)
		assert.Equal(t, err, &UserRoleNotFoundErr{
			RoleID: role.ID,
			UserID: user.ID,
		})
	})

	t.Run("with non-existent user role", func(t *testing.T) {
		roleID := int32(1234)
		userID := int32(4321)

		err := store.Delete(ctx, DeleteUserRoleOpts{
			RoleID: roleID,
			UserID: userID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to delete user role")
	})
}

func TestUserRoleGetByRoleID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	role := createTestRoleForUserRole(ctx, "TESTROLE", t, db)

	totalUsersWithRole := 10
	for i := 1; i <= totalUsersWithRole; i++ {
		username := fmt.Sprintf("ANOTHERTESTUSER%d", i)
		user := createTestUserForUserRole(ctx, fmt.Sprintf("testa%d@example.com", i), username, t, db)

		_, err := store.Create(ctx, CreateUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("missing role id", func(t *testing.T) {
		urs, err := store.GetByRoleID(ctx, GetUserRoleOpts{})
		assert.Error(t, err)
		assert.Nil(t, urs)
		assert.Equal(t, err.Error(), "missing id from sql query")
	})

	t.Run("with provided role id", func(t *testing.T) {
		urs, err := store.GetByRoleID(ctx, GetUserRoleOpts{
			RoleID: role.ID,
		})

		assert.NoError(t, err)
		assert.Len(t, urs, totalUsersWithRole)

		for _, ur := range urs {
			assert.Equal(t, ur.RoleID, role.ID)
		}
	})
}

func TestUserRoleGetByUserID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user := createTestUserForUserRole(ctx, "testuser@example.com", "ANOTHERTESTUSER", t, db)

	totalRoles := 3
	for i := 1; i <= totalRoles; i++ {
		name := fmt.Sprintf("TESTROLE%d", i)
		role := createTestRoleForUserRole(ctx, name, t, db)

		_, err := store.Create(ctx, CreateUserRoleOpts{
			RoleID: role.ID,
			UserID: user.ID,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("missing user id", func(t *testing.T) {
		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{})
		assert.Error(t, err)
		assert.Nil(t, urs)
		assert.Equal(t, err.Error(), "missing user id")
	})

	t.Run("with provided role id", func(t *testing.T) {
		urs, err := store.GetByUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
		})

		assert.NoError(t, err)
		assert.Len(t, urs, totalRoles)

		for _, ur := range urs {
			assert.Equal(t, ur.UserID, user.ID)
		}
	})
}

func TestUserRoleGetByRoleIDAndUserID(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.UserRoles()

	user, role := createUserAndRole(ctx, t, db)
	_, err := store.Create(ctx, CreateUserRoleOpts{
		RoleID: role.ID,
		UserID: user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("without user id", func(t *testing.T) {
		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			RoleID: role.ID,
		})
		assert.Nil(t, ur)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing user id")
	})

	t.Run("without role id", func(t *testing.T) {
		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
		})
		assert.Nil(t, ur)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "missing role id")
	})

	t.Run("with correct args", func(t *testing.T) {
		ur, err := store.GetByRoleIDAndUserID(ctx, GetUserRoleOpts{
			UserID: user.ID,
			RoleID: role.ID,
		})

		assert.NoError(t, err)
		assert.Equal(t, ur.RoleID, role.ID)
		assert.Equal(t, ur.UserID, ur.UserID)
	})
}

func createUserAndRole(ctx context.Context, t *testing.T, db DB) (*types.User, *types.Role) {
	t.Helper()
	user := createTestUserForUserRole(ctx, "a1@example.com", "u1", t, db)
	role := createTestRoleForUserRole(ctx, "ANOTHERTESTROLE - 1", t, db)
	return user, role
}

func createTestRoleForUserRole(ctx context.Context, name string, t *testing.T, db DB) *types.Role {
	t.Helper()
	role, err := db.Roles().Create(ctx, name, false)
	if err != nil {
		t.Fatal(err)
	}
	return role
}

func createTestUserForUserRole(ctx context.Context, email, username string, t *testing.T, db DB) *types.User {
	t.Helper()
	user, err := db.Users().Create(ctx, NewUser{
		Email:                 email,
		Username:              username,
		Password:              "p1",
		EmailVerificationCode: email + username,
	})
	if err != nil {
		t.Fatal(err)
	}
	return user
}
