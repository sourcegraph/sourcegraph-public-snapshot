package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
)

func TestPermissionsResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))

	admin := createTestUser(t, db, true)
	user := createTestUser(t, db, false)

	adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))
	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))

	s, err := NewSchemaWithoutResolvers(db)
	if err != nil {
		t.Fatal(err)
	}

	ps, err := db.Permissions().BulkCreate(ctx, []database.CreatePermissionOpts{
		{
			Namespace: rtypes.BatchChangesNamespace,
			Action:    "READ",
		},
		{
			Namespace: rtypes.BatchChangesNamespace,
			Action:    "WRITE",
		},
		{
			Namespace: rtypes.BatchChangesNamespace,
			Action:    "EXECUTE",
		},
	})
	require.NoError(t, err)

	t.Run("as non site-administrator", func(t *testing.T) {
		input := map[string]any{"first": 1}
		var response struct{ Permissions apitest.PermissionConnection }
		errs := apitest.Exec(actor.WithActor(userCtx, actor.FromUser(user.ID)), t, s, input, &response, queryPermissionConnection)

		require.Len(t, errs, 1)
		require.Equal(t, errs[0].Message, "must be site admin")
	})

	t.Run("as site-administrator", func(t *testing.T) {
		want := []apitest.Permission{
			{
				ID: string(MarshalPermissionID(ps[2].ID)),
			},
			{
				ID: string(MarshalPermissionID(ps[1].ID)),
			},
			{
				ID: string(MarshalPermissionID(ps[0].ID)),
			},
		}

		tests := []struct {
			firstParam          int
			wantHasPreviousPage bool
			wantHasNextPage     bool
			wantTotalCount      int
			wantNodes           []apitest.Permission
		}{
			{firstParam: 1, wantHasNextPage: true, wantHasPreviousPage: false, wantTotalCount: 3, wantNodes: want[:1]},
			{firstParam: 2, wantHasNextPage: true, wantHasPreviousPage: false, wantTotalCount: 3, wantNodes: want[:2]},
			{firstParam: 3, wantHasNextPage: false, wantHasPreviousPage: false, wantTotalCount: 3, wantNodes: want},
			{firstParam: 4, wantHasNextPage: false, wantHasPreviousPage: false, wantTotalCount: 3, wantNodes: want},
		}

		for _, tc := range tests {
			t.Run(fmt.Sprintf("first=%d", tc.firstParam), func(t *testing.T) {
				input := map[string]any{"first": int64(tc.firstParam)}
				var response struct{ Permissions apitest.PermissionConnection }
				apitest.MustExec(actor.WithActor(adminCtx, actor.FromUser(admin.ID)), t, s, input, &response, queryPermissionConnection)

				wantConnection := apitest.PermissionConnection{
					TotalCount: tc.wantTotalCount,
					PageInfo: apitest.PageInfo{
						HasNextPage:     tc.wantHasNextPage,
						EndCursor:       response.Permissions.PageInfo.EndCursor,
						HasPreviousPage: tc.wantHasPreviousPage,
					},
					Nodes: tc.wantNodes,
				}

				if diff := cmp.Diff(wantConnection, response.Permissions); diff != "" {
					t.Fatalf("wrong permissions response (-want +got):\n%s", diff)
				}
			})
		}
	})
}

const queryPermissionConnection = `
query($first: Int!) {
	permissions(first: $first) {
		totalCount
		pageInfo {
			hasNextPage
			endCursor
		}
		nodes {
			id
		}
	}
}
`

// Check if its a different user, site admin and same user
func TestUserPermissionsListing(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))

	userID := createTestUser(t, db, false).ID
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))

	adminUserID := createTestUser(t, db, true).ID
	adminActorCtx := actor.WithActor(ctx, actor.FromUser(adminUserID))

	s, err := NewSchemaWithoutResolvers(db)
	require.NoError(t, err)

	// create a new role
	role, err := db.Roles().Create(ctx, "TEST-ROLE", false)
	require.NoError(t, err)

	err = db.UserRoles().Assign(ctx, database.AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: userID,
	})
	require.NoError(t, err)

	p, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: rtypes.BatchChangesNamespace,
		Action:    "READ",
	})
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, database.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: p.ID,
	})
	require.NoError(t, err)

	t.Run("listing a user's permissions (same user)", func(t *testing.T) {
		userAPIID := string(MarshalUserID(userID))
		input := map[string]any{"node": userAPIID}

		want := apitest.User{
			ID: userAPIID,
			Permissions: apitest.PermissionConnection{
				TotalCount: 1,
				Nodes: []apitest.Permission{
					{
						ID: string(MarshalPermissionID(p.ID)),
					},
				},
			},
		}

		var response struct{ Node apitest.User }
		apitest.MustExec(actorCtx, t, s, input, &response, listUserPermissions)

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("wrong permission response (-want +got):\n%s", diff)
		}
	})

	t.Run("listing a user's permissions (site admin)", func(t *testing.T) {
		userAPIID := string(MarshalUserID(userID))
		input := map[string]any{"node": userAPIID}

		want := apitest.User{
			ID: userAPIID,
			Permissions: apitest.PermissionConnection{
				TotalCount: 1,
				Nodes: []apitest.Permission{
					{
						ID: string(MarshalPermissionID(p.ID)),
					},
				},
			},
		}

		var response struct{ Node apitest.User }
		apitest.MustExec(adminActorCtx, t, s, input, &response, listUserPermissions)

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("wrong permissions response (-want +got):\n%s", diff)
		}
	})

	t.Run("non site-admin listing another user's permission", func(t *testing.T) {
		userAPIID := string(MarshalUserID(adminUserID))
		input := map[string]any{"node": userAPIID}

		var response struct{}
		errs := apitest.Exec(actorCtx, t, s, input, &response, listUserPermissions)
		require.Len(t, errs, 1)
		require.Equal(t, auth.ErrMustBeSiteAdminOrSameUser.Error(), errs[0].Message)
	})
}

const listUserPermissions = `
query ($node: ID!) {
	node(id: $node) {
		... on User {
			id
			permissions {
				totalCount
				nodes {
					id
				}
			}
		}
	}
}
`
