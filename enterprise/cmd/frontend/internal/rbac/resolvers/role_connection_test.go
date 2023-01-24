package resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rbac/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
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

	s, err := newSchema(db, &Resolver{logger: logger, db: db})
	if err != nil {
		t.Fatal(err)
	}

	userID := createTestUser(t, db, false).ID

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
		firstParam      int
		wantHasNextPage bool
		wantTotalCount  int
		wantNodes       []apitest.Role
	}{
		{firstParam: 1, wantHasNextPage: true, wantTotalCount: 3, wantNodes: want[:1]},
		{firstParam: 2, wantHasNextPage: true, wantTotalCount: 3, wantNodes: want[:2]},
		{firstParam: 3, wantHasNextPage: false, wantTotalCount: 3, wantNodes: want},
		{firstParam: 4, wantHasNextPage: false, wantTotalCount: 3, wantNodes: want},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("first=%d", tc.firstParam), func(t *testing.T) {
			input := map[string]any{"first": int64(tc.firstParam)}
			var response struct{ Roles apitest.RoleConnection }
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryRoleConnection)

			wantConnection := apitest.RoleConnection{
				TotalCount: tc.wantTotalCount,
				PageInfo: apitest.PageInfo{
					HasNextPage: tc.wantHasNextPage,
					EndCursor:   response.Roles.PageInfo.EndCursor,
				},
				Nodes: tc.wantNodes,
			}

			if diff := cmp.Diff(wantConnection, response.Roles); diff != "" {
				t.Fatalf("wrong roles response (-want +got):\n%s", diff)
			}
		})
	}
}

const queryRoleConnection = `
query($first: Int, $after: String) {
	roles(first: $first, after: $after) {
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

	_, err = db.UserRoles().Create(ctx, database.CreateUserRoleOpts{
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
		t.Skip()
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
}

const listUserRoles = `
query ($node: ID!) {
	node(id: $node) {
		... on User {
			id
			roles {
				totalCount
				nodes {
					id
				}
			}
		}
	}
}
`
