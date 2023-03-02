package resolvers

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rbac/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestPermissionsResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	admin := createTestUser(t, db, true)
	user := createTestUser(t, db, false)

	adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))
	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))

	s, err := newSchema(db, &Resolver{logger: logger, db: db})
	if err != nil {
		t.Fatal(err)
	}

	ps, err := db.Permissions().BulkCreate(ctx, []database.CreatePermissionOpts{
		{
			Namespace: types.BatchChangesNamespace,
			Action:    "READ",
		},
		{
			Namespace: types.BatchChangesNamespace,
			Action:    "WRITE",
		},
		{
			Namespace: types.BatchChangesNamespace,
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
				ID: string(marshalPermissionID(ps[2].ID)),
			},
			{
				ID: string(marshalPermissionID(ps[1].ID)),
			},
			{
				ID: string(marshalPermissionID(ps[0].ID)),
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
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	userID := createTestUser(t, db, false).ID
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))

	adminUserID := createTestUser(t, db, true).ID
	adminActorCtx := actor.WithActor(ctx, actor.FromUser(adminUserID))

	r := &Resolver{logger: logger, db: db}
	s, err := newSchema(db, r)
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
		Namespace: types.BatchChangesNamespace,
		Action:    "READ",
	})
	require.NoError(t, err)

	err = db.RolePermissions().Assign(ctx, database.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: p.ID,
	})
	require.NoError(t, err)

	t.Run("listing a user's permissions (same user)", func(t *testing.T) {
		userAPIID := string(gql.MarshalUserID(userID))
		input := map[string]any{"node": userAPIID}

		want := apitest.User{
			ID: userAPIID,
			Permissions: apitest.PermissionConnection{
				TotalCount: 1,
				Nodes: []apitest.Permission{
					{
						ID: string(marshalPermissionID(p.ID)),
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
		userAPIID := string(gql.MarshalUserID(userID))
		input := map[string]any{"node": userAPIID}

		want := apitest.User{
			ID: userAPIID,
			Permissions: apitest.PermissionConnection{
				TotalCount: 1,
				Nodes: []apitest.Permission{
					{
						ID: string(marshalPermissionID(p.ID)),
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
		userAPIID := string(gql.MarshalUserID(adminUserID))
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
			permissions(first: 10) {
				totalCount
				nodes {
					id
				}
			}
		}
	}
}
`

func TestSetPermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	admin := createTestUser(t, db, true)
	user := createTestUser(t, db, false)

	adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))
	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))

	s, err := newSchema(db, &Resolver{logger: logger, db: db})
	if err != nil {
		t.Fatal(err)
	}

	permissions := createPermissions(ctx, t, db)

	roleWithoutPermissions := createRoleWithPermissions(ctx, t, db, "test-role")
	roleWithAllPermissions := createRoleWithPermissions(ctx, t, db, "test-role-1", permissions...)
	roleWithAllPermissions2 := createRoleWithPermissions(ctx, t, db, "test-role-2", permissions...)
	roleWithOnePermission := createRoleWithPermissions(ctx, t, db, "test-role-3", permissions[0])

	var permissionIDs []graphql.ID
	for _, p := range permissions {
		permissionIDs = append(permissionIDs, marshalPermissionID(p.ID))
	}

	t.Run("as non-site-admin", func(t *testing.T) {
		input := map[string]any{"role": marshalRoleID(roleWithoutPermissions.ID), "permissions": []int32{}}
		var response struct{ Permissions apitest.EmptyResponse }
		errs := apitest.Exec(userCtx, t, s, input, &response, setPermissionsQuery)

		require.Len(t, errs, 1)
		require.ErrorContains(t, errs[0], "must be site admin")
	})

	t.Run("as site-admin", func(t *testing.T) {
		// There are no permissions assigned to `roleWithoutPermissions`, so we asssign all permissions to that role.
		// Passing an array of permissions will assign the permissions to the role.
		t.Run("assign permissions", func(t *testing.T) {
			input := map[string]any{"role": marshalRoleID(roleWithoutPermissions.ID), "permissions": permissionIDs}
			var response struct{ Permissions apitest.EmptyResponse }
			apitest.MustExec(adminCtx, t, s, input, &response, setPermissionsQuery)

			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithoutPermissions.ID)
			require.Len(t, rps, len(permissionIDs))

			sort.Slice(rps, func(i, j int) bool {
				return rps[i].PermissionID < rps[j].PermissionID
			})
			for index, rp := range rps {
				require.Equal(t, rp.RoleID, roleWithoutPermissions.ID)
				require.Equal(t, rp.PermissionID, permissions[index].ID)
			}
		})

		t.Run("revoke permissions", func(t *testing.T) {
			input := map[string]any{"role": marshalRoleID(roleWithAllPermissions.ID), "permissions": []graphql.ID{}}
			var response struct{ Permissions apitest.EmptyResponse }
			apitest.MustExec(adminCtx, t, s, input, &response, setPermissionsQuery)

			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithAllPermissions.ID)
			require.Len(t, rps, 0)
		})

		t.Run("assign and revoke permissions", func(t *testing.T) {
			// omitting the first permissions (which is already assigned to the role) will revoke it for the role.
			input := map[string]any{"role": marshalRoleID(roleWithOnePermission.ID), "permissions": permissionIDs[1:]}
			var response struct{ Permissions apitest.EmptyResponse }
			apitest.MustExec(adminCtx, t, s, input, &response, setPermissionsQuery)

			// Since this role has the first permission assigned to it, since we
			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithOnePermission.ID)
			require.Len(t, rps, 2)

			sort.Slice(rps, func(i, j int) bool {
				return rps[i].PermissionID < rps[j].PermissionID
			})
			for index, rp := range rps {
				require.Equal(t, rp.RoleID, roleWithOnePermission.ID)
				require.Equal(t, rp.PermissionID, permissions[index+1].ID)
			}
		})

		t.Run("no change", func(t *testing.T) {
			input := map[string]any{"role": marshalRoleID(roleWithAllPermissions2.ID), "permissions": permissionIDs}
			var response struct{ Permissions apitest.EmptyResponse }
			apitest.MustExec(adminCtx, t, s, input, &response, setPermissionsQuery)

			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithAllPermissions2.ID)
			require.Len(t, rps, len(permissions))

			sort.Slice(rps, func(i, j int) bool {
				return rps[i].PermissionID < rps[j].PermissionID
			})
			for index, rp := range rps {
				require.Equal(t, rp.RoleID, roleWithAllPermissions2.ID)
				require.Equal(t, rp.PermissionID, permissions[index].ID)
			}
		})
	})
}

const setPermissionsQuery = `
mutation($role: ID!, $permissions: [ID!]!) {
	setPermissions(role: $role, permissions: $permissions) {
		alwaysNil
	}
}
`

func createPermissions(ctx context.Context, t *testing.T, db database.DB) []*types.Permission {
	t.Helper()

	ps, err := db.Permissions().BulkCreate(ctx, []database.CreatePermissionOpts{
		{
			Namespace: types.BatchChangesNamespace,
			Action:    "READ",
		},
		{
			Namespace: types.BatchChangesNamespace,
			Action:    "WRITE",
		},
		{
			Namespace: types.BatchChangesNamespace,
			Action:    "EXECUTE",
		},
	})
	require.NoError(t, err)
	return ps
}

func createRoleWithPermissions(ctx context.Context, t *testing.T, db database.DB, name string, permissions ...*types.Permission) *types.Role {
	t.Helper()

	role, err := db.Roles().Create(ctx, name, false)
	require.NoError(t, err)

	if len(permissions) > 0 {
		var opts = database.BulkAssignPermissionsToRoleOpts{RoleID: role.ID}
		for _, permission := range permissions {
			opts.Permissions = append(opts.Permissions, permission.ID)
		}

		err = db.RolePermissions().BulkAssignPermissionsToRole(ctx, opts)
		require.NoError(t, err)
	}

	return role
}

func getPermissionsAssignedToRole(ctx context.Context, t *testing.T, db database.DB, roleID int32) []*types.RolePermission {
	t.Helper()

	rps, err := db.RolePermissions().GetByRoleID(ctx, database.GetRolePermissionOpts{RoleID: roleID})
	require.NoError(t, err)

	return rps
}
