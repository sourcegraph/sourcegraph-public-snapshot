package resolvers

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rbac/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

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
