package resolvers

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rbac/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRoleConnectionResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	userID := createTestUser(t, db, false).ID
	userCtx := actor.WithActor(ctx, actor.FromUser(userID))

	adminID := createTestUser(t, db, true).ID
	adminCtx := actor.WithActor(ctx, actor.FromUser(adminID))

	s, err := newSchema(db, &Resolver{logger: logger, db: db})
	if err != nil {
		t.Fatal(err)
	}

	// All sourcegraph instances are seeded with two system roles at migration,
	// so we take those into account when querying roles.
	siteAdminRole, err := db.Roles().Get(ctx, database.GetRoleOpts{
		Name: string(types.SiteAdministratorSystemRole),
	})
	assert.NoError(t, err)

	userRole, err := db.Roles().Get(ctx, database.GetRoleOpts{
		Name: string(types.UserSystemRole),
	})
	assert.NoError(t, err)

	r, err := db.Roles().Create(ctx, "TEST-ROLE", false)
	assert.NoError(t, err)

	t.Run("as non site-administrator", func(t *testing.T) {
		input := map[string]any{"first": 1}
		var response struct{ Permissions apitest.PermissionConnection }
		errs := apitest.Exec(userCtx, t, s, input, &response, queryPermissionConnection)

		assert.Len(t, errs, 1)
		assert.Equal(t, errs[0].Message, "must be site admin")
	})

	t.Run("as site-administrator", func(t *testing.T) {
		want := []apitest.Role{
			{
				ID: string(marshalRoleID(userRole.ID)),
			},
			{
				ID: string(marshalRoleID(siteAdminRole.ID)),
			},
			{
				ID: string(marshalRoleID(r.ID)),
			},
		}

		tests := []struct {
			firstParam          int
			wantHasNextPage     bool
			wantHasPreviousPage bool
			wantTotalCount      int
			wantNodes           []apitest.Role
		}{
			{firstParam: 1, wantHasNextPage: true, wantHasPreviousPage: false, wantTotalCount: 3, wantNodes: want[:1]},
			{firstParam: 2, wantHasNextPage: true, wantHasPreviousPage: false, wantTotalCount: 3, wantNodes: want[:2]},
			{firstParam: 3, wantHasNextPage: false, wantHasPreviousPage: false, wantTotalCount: 3, wantNodes: want},
			{firstParam: 4, wantHasNextPage: false, wantHasPreviousPage: false, wantTotalCount: 3, wantNodes: want},
		}

		for _, tc := range tests {
			t.Run(fmt.Sprintf("first=%d", tc.firstParam), func(t *testing.T) {
				input := map[string]any{"first": int64(tc.firstParam)}
				var response struct{ Roles apitest.RoleConnection }
				apitest.MustExec(adminCtx, t, s, input, &response, queryRoleConnection)

				wantConnection := apitest.RoleConnection{
					TotalCount: tc.wantTotalCount,
					PageInfo: apitest.PageInfo{
						HasNextPage:     tc.wantHasNextPage,
						HasPreviousPage: tc.wantHasPreviousPage,
					},
					Nodes: tc.wantNodes,
				}

				if diff := cmp.Diff(wantConnection, response.Roles); diff != "" {
					t.Fatalf("wrong roles response (-want +got):\n%s", diff)
				}
			})
		}
	})
}

const queryRoleConnection = `
query($first: Int!) {
	roles(first: $first) {
		totalCount
		pageInfo {
			hasNextPage
			hasPreviousPage
		}
		nodes {
			id
		}
	}
}
`

func TestUserRoleListing(t *testing.T) {
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
	assert.NoError(t, err)

	// create a new role
	role, err := db.Roles().Create(ctx, "TEST-ROLE", false)
	assert.NoError(t, err)

	err = db.UserRoles().Assign(ctx, database.AssignUserRoleOpts{
		RoleID: role.ID,
		UserID: userID,
	})
	assert.NoError(t, err)

	t.Run("listing a user's roles (same user)", func(t *testing.T) {
		userAPIID := string(gql.MarshalUserID(userID))
		input := map[string]any{"node": userAPIID}

		want := apitest.User{
			ID: userAPIID,
			Roles: apitest.RoleConnection{
				TotalCount: 1,
				Nodes: []apitest.Role{
					{
						ID: string(marshalRoleID(role.ID)),
					},
				},
			},
		}

		var response struct{ Node apitest.User }
		apitest.MustExec(actorCtx, t, s, input, &response, listUserRoles)

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("wrong role response (-want +got):\n%s", diff)
		}
	})

	t.Run("listing a user's roles (site admin)", func(t *testing.T) {
		userAPIID := string(gql.MarshalUserID(userID))
		input := map[string]any{"node": userAPIID}

		want := apitest.User{
			ID: userAPIID,
			Roles: apitest.RoleConnection{
				TotalCount: 1,
				Nodes: []apitest.Role{
					{
						ID: string(marshalRoleID(role.ID)),
					},
				},
			},
		}

		var response struct{ Node apitest.User }
		apitest.MustExec(adminActorCtx, t, s, input, &response, listUserRoles)

		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("wrong roles response (-want +got):\n%s", diff)
		}
	})

	t.Run("non site-admin listing another user's roles", func(t *testing.T) {
		userAPIID := string(gql.MarshalUserID(adminUserID))
		input := map[string]any{"node": userAPIID}

		var response struct{}
		errs := apitest.Exec(actorCtx, t, s, input, &response, listUserRoles)
		assert.Len(t, errs, 1)
		assert.Equal(t, auth.ErrMustBeSiteAdminOrSameUser.Error(), errs[0].Message)
	})
}

const listUserRoles = `
query ($node: ID!) {
	node(id: $node) {
		... on User {
			id
			roles(first: 50) {
				totalCount
				nodes {
					id
				}
			}
		}
	}
}
`

func TestDeleteRole(t *testing.T) {
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
	assert.NoError(t, err)

	// create a new role
	role, err := db.Roles().Create(ctx, "TEST-ROLE", false)
	assert.NoError(t, err)

	t.Run("as non site-admin", func(t *testing.T) {
		roleID := string(marshalRoleID(role.ID))
		input := map[string]any{"role": roleID}

		var response struct{ DeleteRole apitest.EmptyResponse }
		errs := apitest.Exec(actorCtx, t, s, input, &response, deleteRoleMutation)

		if len(errs) != 1 {
			t.Fatalf("expected single errors, but got %d", len(errs))
		}
		if have, want := errs[0].Message, "must be site admin"; have != want {
			t.Fatalf("wrong error. want=%q, have=%q", want, have)
		}
	})

	t.Run("as site-admin", func(t *testing.T) {
		roleID := string(marshalRoleID(role.ID))
		input := map[string]any{"role": roleID}

		var response struct{ DeleteRole apitest.EmptyResponse }

		// First time it should work, because the role exists
		apitest.MustExec(adminActorCtx, t, s, input, &response, deleteRoleMutation)

		// Second time it should fail
		errs := apitest.Exec(adminActorCtx, t, s, input, &response, deleteRoleMutation)

		if len(errs) != 1 {
			t.Fatalf("expected a single error, but got %d", len(errs))
		}
		if have, want := errs[0].Message, fmt.Sprintf("failed to delete role: role with ID %d not found", role.ID); have != want {
			t.Fatalf("wrong error code. want=%q, have=%q", want, have)
		}
	})
}

const deleteRoleMutation = `
mutation DeleteRole($role: ID!) {
	deleteRole(role: $role) {
		alwaysNil
	}
}
`

func TestCreateRole(t *testing.T) {
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
	assert.NoError(t, err)

	perm, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: types.BatchChangesNamespace,
		Action:    "READ",
	})
	require.NoError(t, err)

	t.Run("as non site-admin", func(t *testing.T) {
		input := map[string]any{"name": "TEST-ROLE", "permissions": []graphql.ID{}}

		var response struct{ CreateRole apitest.Role }
		errs := apitest.Exec(actorCtx, t, s, input, &response, createRoleMutation)

		if len(errs) != 1 {
			t.Fatalf("expected a single error, but got %d", len(errs))
		}
		if have, want := errs[0].Message, "must be site admin"; have != want {
			t.Fatalf("wrong error. want=%q, have=%q", want, have)
		}
	})

	t.Run("as site-admin", func(t *testing.T) {
		t.Run("without permissions", func(t *testing.T) {
			input := map[string]any{"name": "TEST-ROLE-1", "permissions": []graphql.ID{}}

			var response struct{ CreateRole apitest.Role }
			// First time it should work, because the role exists
			apitest.MustExec(adminActorCtx, t, s, input, &response, createRoleMutation)

			// Second time it should fail because role names must be unique
			errs := apitest.Exec(adminActorCtx, t, s, input, &response, createRoleMutation)
			if len(errs) != 1 {
				t.Fatalf("expected a single error, but got %d", len(errs))
			}
			if have, want := errs[0].Message, "cannot create role: err_name_exists"; have != want {
				t.Fatalf("wrong error code. want=%q, have=%q", want, have)
			}

			roleID, err := unmarshalRoleID(graphql.ID(response.CreateRole.ID))
			require.NoError(t, err)
			rps, err := db.RolePermissions().GetByRoleID(ctx, database.GetRolePermissionOpts{
				RoleID: roleID,
			})
			require.NoError(t, err)
			require.Len(t, rps, 0)
		})

		t.Run("with permissions", func(t *testing.T) {
			input := map[string]any{"name": "TEST-ROLE-2", "permissions": []graphql.ID{
				marshalPermissionID(perm.ID),
			}}

			var response struct{ CreateRole apitest.Role }
			// First time it should work, because the role exists
			apitest.MustExec(adminActorCtx, t, s, input, &response, createRoleMutation)

			// Second time it should fail because role names must be unique
			errs := apitest.Exec(adminActorCtx, t, s, input, &response, createRoleMutation)
			if len(errs) != 1 {
				t.Fatalf("expected a single error, but got %d", len(errs))
			}
			if have, want := errs[0].Message, "cannot create role: err_name_exists"; have != want {
				t.Fatalf("wrong error code. want=%q, have=%q", want, have)
			}

			roleID, err := unmarshalRoleID(graphql.ID(response.CreateRole.ID))
			require.NoError(t, err)
			rps, err := db.RolePermissions().GetByRoleID(ctx, database.GetRolePermissionOpts{
				RoleID: roleID,
			})
			require.NoError(t, err)
			require.Len(t, rps, 1)
			require.Equal(t, rps[0].PermissionID, perm.ID)
		})
	})
}

const createRoleMutation = `
mutation CreateRole($name: String!, $permissions: [ID!]!) {
	createRole(name: $name, permissions: $permissions) {
		id
		name
		system
	}
}
`

func TestSetRoles(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	uID := createTestUser(t, db, false).ID
	userID := gql.MarshalUserID(uID)
	userCtx := actor.WithActor(ctx, actor.FromUser(uID))

	aID := createTestUser(t, db, true).ID
	adminID := gql.MarshalUserID(aID)
	adminCtx := actor.WithActor(ctx, actor.FromUser(aID))

	s, err := newSchema(db, &Resolver{logger: logger, db: db})
	if err != nil {
		t.Fatal(err)
	}

	r1, err := db.Roles().Create(ctx, "TEST-ROLE-1", false)
	require.NoError(t, err)

	r2, err := db.Roles().Create(ctx, "TEST-ROLE-2", false)
	require.NoError(t, err)

	r3, err := db.Roles().Create(ctx, "TEST-ROLE-3", false)
	require.NoError(t, err)

	var marshalledRoles = []graphql.ID{
		marshalRoleID(r1.ID),
		marshalRoleID(r2.ID),
		marshalRoleID(r3.ID),
	}

	var roles = []*types.Role{r1, r2, r3}

	userWithoutARole := createUserWithRoles(ctx, t, db)
	userWithAllRoles := createUserWithRoles(ctx, t, db, roles...)
	userWithAllRoles2 := createUserWithRoles(ctx, t, db, roles...)
	userWithOneRole := createUserWithRoles(ctx, t, db, roles[0])

	t.Run("as non site-admin", func(t *testing.T) {
		input := map[string]any{"user": userID, "roles": marshalledRoles}
		var response struct{ Se apitest.EmptyResponse }
		errs := apitest.Exec(userCtx, t, s, input, &response, setRolesQuery)

		require.Len(t, errs, 1)
		require.ErrorContains(t, errs[0], "must be site admin")
	})

	t.Run("as self", func(t *testing.T) {
		input := map[string]any{"user": adminID, "roles": marshalledRoles}
		var response struct{ AssignRolesToUser apitest.EmptyResponse }
		errs := apitest.Exec(adminCtx, t, s, input, &response, setRolesQuery)

		require.Len(t, errs, 1)
		require.ErrorContains(t, errs[0], "cannot assign role to self")
	})

	t.Run("as site-admin", func(t *testing.T) {
		// There are no permissions assigned to `userWithoutARole`, so we asssign all roles to that user.
		// Passing a slice of roles will assign the roles to the user.
		t.Run("assign roles", func(t *testing.T) {
			input := map[string]any{"user": gql.MarshalUserID(userWithoutARole.ID), "roles": marshalledRoles}
			var response struct{ SetRoles apitest.EmptyResponse }

			apitest.MustExec(adminCtx, t, s, input, &response, setRolesQuery)

			urs, err := db.UserRoles().GetByUserID(ctx, database.GetUserRoleOpts{
				UserID: userWithoutARole.ID,
			})
			require.NoError(t, err)

			sort.Slice(urs, func(i, j int) bool {
				return urs[i].RoleID < urs[j].RoleID
			})
			for index, ur := range urs {
				require.Equal(t, ur.RoleID, roles[index].ID)
				require.Equal(t, ur.UserID, userWithoutARole.ID)
			}
		})

		// We pass an empty role slice to revoke all roles assigned to the user.
		t.Run("revoke roles", func(t *testing.T) {
			input := map[string]any{"user": gql.MarshalUserID(userWithAllRoles.ID), "roles": []graphql.ID{}}
			var response struct{ SetRoles apitest.EmptyResponse }

			apitest.MustExec(adminCtx, t, s, input, &response, setRolesQuery)

			urs, err := db.UserRoles().GetByUserID(ctx, database.GetUserRoleOpts{
				UserID: userWithAllRoles.ID,
			})
			require.NoError(t, err)
			require.Len(t, urs, 0)
		})

		t.Run("assign and revoke permissions", func(t *testing.T) {
			// omitting the first role (which is already assigned to the user) will revoke it for the user.
			input := map[string]any{"roles": marshalledRoles[1:], "user": gql.MarshalUserID(userWithOneRole.ID)}
			var response struct{ SetRoles apitest.EmptyResponse }
			apitest.MustExec(adminCtx, t, s, input, &response, setRolesQuery)

			urs, err := db.UserRoles().GetByUserID(ctx, database.GetUserRoleOpts{
				UserID: userWithOneRole.ID,
			})
			require.NoError(t, err)
			require.Len(t, urs, len(roles)-1)

			sort.Slice(urs, func(i, j int) bool {
				return urs[i].RoleID < urs[j].RoleID
			})
			for index, ur := range urs {
				require.Equal(t, ur.RoleID, roles[index+1].ID)
				require.Equal(t, ur.UserID, userWithOneRole.ID)
			}
		})

		t.Run("no change", func(t *testing.T) {
			input := map[string]any{"user": gql.MarshalUserID(userWithAllRoles2.ID), "roles": marshalledRoles}
			var response struct{ SetRoles apitest.EmptyResponse }

			apitest.MustExec(adminCtx, t, s, input, &response, setRolesQuery)

			urs, err := db.UserRoles().GetByUserID(ctx, database.GetUserRoleOpts{
				UserID: userWithAllRoles2.ID,
			})
			require.NoError(t, err)
			require.Len(t, urs, len(roles))

			sort.Slice(urs, func(i, j int) bool {
				return urs[i].RoleID < urs[j].RoleID
			})
			for index, ur := range urs {
				require.Equal(t, ur.RoleID, roles[index].ID)
				require.Equal(t, ur.UserID, userWithAllRoles2.ID)
			}
		})
	})
}

const setRolesQuery = `
mutation SetRoles($roles: [ID!]!, $user: ID!) {
	setRoles(roles: $roles, user: $user) {
		alwaysNil
	}
}
`

func createUserWithRoles(ctx context.Context, t *testing.T, db database.DB, roles ...*types.Role) *types.User {
	t.Helper()

	user := createTestUser(t, db, false)

	if len(roles) > 0 {
		var opts = database.BulkAssignRolesToUserOpts{UserID: user.ID}
		for _, role := range roles {
			opts.Roles = append(opts.Roles, role.ID)
		}

		err := db.UserRoles().BulkAssignRolesToUser(ctx, opts)
		require.NoError(t, err)
	}

	return user
}
