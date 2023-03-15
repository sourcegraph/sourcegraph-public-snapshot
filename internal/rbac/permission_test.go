package rbac

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCheckCurrentUserHasPermission(t *testing.T) {
	ctx := context.Background()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	newUser1 := database.NewUser{Username: "username-1"}
	u1, err := db.Users().Create(ctx, newUser1)
	require.NoError(t, err)

	newUser2 := database.NewUser{Username: "username-2"}
	u2, err := db.Users().Create(ctx, newUser2)
	require.NoError(t, err)

	p, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: types.BatchChangesNamespace,
		Action:    "EXECUTE",
	})
	require.NoError(t, err)

	r, err := db.Roles().Create(ctx, "TEST-ROLE", false)
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, database.AssignRolePermissionOpts{
		RoleID:       r.ID,
		PermissionID: p.ID,
	})
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, database.AssignUserRoleOpts{
		UserID: u2.ID,
		RoleID: r.ID,
	})
	require.NoError(t, err)

	tests := []struct {
		name       string
		context    context.Context
		permission string

		expectedErr error
	}{
		{
			name:        "internal actor",
			context:     actor.WithInternalActor(ctx),
			permission:  "",
			expectedErr: nil,
		},
		{
			name:        "non-existent actor",
			context:     actor.WithActor(ctx, &actor.Actor{UID: 9389}),
			permission:  "",
			expectedErr: auth.ErrNotAuthenticated,
		},
		{
			name:        "invalid permission",
			context:     actor.WithActor(ctx, &actor.Actor{UID: u1.ID}),
			permission:  "BATCH_CHANGE@EXEC",
			expectedErr: invalidPermissionDisplayName,
		},
		{
			name:        "unauthorized user",
			context:     actor.WithActor(ctx, &actor.Actor{UID: u1.ID}),
			permission:  p.DisplayName(),
			expectedErr: &ErrNotAuthorized{Permission: p.DisplayName()},
		},
		{
			name:        "authorized user",
			context:     actor.WithActor(ctx, &actor.Actor{UID: u2.ID}),
			permission:  p.DisplayName(),
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckCurrentUserHasPermission(tc.context, db, tc.permission)
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
