package resolvers

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/rbac/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestDeleteRole(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))

	userID := createTestUser(t, db, false).ID
	actorCtx := actor.WithActor(ctx, actor.FromMockUser(userID))

	adminUserID := createTestUser(t, db, true).ID
	adminActorCtx := actor.WithActor(ctx, actor.FromMockUser(adminUserID))

	r := &Resolver{logger: logger, db: db}
	s, err := newSchema(db, r)
	require.NoError(t, err)

	// create a new role
	role, err := db.Roles().Create(ctx, "TEST-ROLE", false)
	require.NoError(t, err)

	t.Run("as non site-admin", func(t *testing.T) {
		roleID := string(gql.MarshalRoleID(role.ID))
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
		roleID := string(gql.MarshalRoleID(role.ID))
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
	db := database.NewDB(logger, dbtest.NewDB(t))

	userID := createTestUser(t, db, false).ID
	actorCtx := actor.WithActor(ctx, actor.FromMockUser(userID))

	adminUserID := createTestUser(t, db, true).ID
	adminActorCtx := actor.WithActor(ctx, actor.FromMockUser(adminUserID))

	r := &Resolver{logger: logger, db: db}
	s, err := newSchema(db, r)
	require.NoError(t, err)

	perm, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: rtypes.BatchChangesNamespace,
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

			roleID, err := gql.UnmarshalRoleID(graphql.ID(response.CreateRole.ID))
			require.NoError(t, err)
			rps, err := db.RolePermissions().GetByRoleID(ctx, database.GetRolePermissionOpts{
				RoleID: roleID,
			})
			require.NoError(t, err)
			require.Len(t, rps, 0)
		})

		t.Run("with permissions", func(t *testing.T) {
			input := map[string]any{"name": "TEST-ROLE-2", "permissions": []graphql.ID{
				gql.MarshalPermissionID(perm.ID),
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

			roleID, err := gql.UnmarshalRoleID(graphql.ID(response.CreateRole.ID))
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

func TestSetPermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))

	admin := createTestUser(t, db, true)
	user := createTestUser(t, db, false)

	adminCtx := actor.WithActor(ctx, actor.FromMockUser(admin.ID))
	userCtx := actor.WithActor(ctx, actor.FromMockUser(user.ID))

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
		permissionIDs = append(permissionIDs, gql.MarshalPermissionID(p.ID))
	}

	t.Run("as non-site-admin", func(t *testing.T) {
		input := map[string]any{"role": gql.MarshalRoleID(roleWithoutPermissions.ID), "permissions": []int32{}}
		var response struct{ Permissions apitest.EmptyResponse }
		errs := apitest.Exec(userCtx, t, s, input, &response, setPermissionsQuery)

		require.Len(t, errs, 1)
		require.ErrorContains(t, errs[0], "must be site admin")
	})

	t.Run("as site-admin", func(t *testing.T) {
		// There are no permissions assigned to `roleWithoutPermissions`, so we asssign all permissions to that role.
		// Passing an array of permissions will assign the permissions to the role.
		t.Run("assign permissions", func(t *testing.T) {
			input := map[string]any{"role": gql.MarshalRoleID(roleWithoutPermissions.ID), "permissions": permissionIDs}
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
			input := map[string]any{"role": gql.MarshalRoleID(roleWithAllPermissions.ID), "permissions": []graphql.ID{}}
			var response struct{ Permissions apitest.EmptyResponse }
			apitest.MustExec(adminCtx, t, s, input, &response, setPermissionsQuery)

			rps := getPermissionsAssignedToRole(ctx, t, db, roleWithAllPermissions.ID)
			require.Len(t, rps, 0)
		})

		t.Run("assign and revoke permissions", func(t *testing.T) {
			// omitting the first permissions (which is already assigned to the role) will revoke it for the role.
			input := map[string]any{"role": gql.MarshalRoleID(roleWithOnePermission.ID), "permissions": permissionIDs[1:]}
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
			input := map[string]any{"role": gql.MarshalRoleID(roleWithAllPermissions2.ID), "permissions": permissionIDs}
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

func TestSetRoles(t *testing.T) {
	logger := logtest.Scoped(t)
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))

	uID := createTestUser(t, db, false).ID
	userID := gql.MarshalUserID(uID)
	userCtx := actor.WithActor(ctx, actor.FromMockUser(uID))

	aID := createTestUser(t, db, true).ID
	adminCtx := actor.WithActor(ctx, actor.FromMockUser(aID))

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
		gql.MarshalRoleID(r1.ID),
		gql.MarshalRoleID(r2.ID),
		gql.MarshalRoleID(r3.ID),
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

		t.Run("assign and revoke roles", func(t *testing.T) {
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

func createPermissions(ctx context.Context, t *testing.T, db database.DB) []*types.Permission {
	t.Helper()

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
