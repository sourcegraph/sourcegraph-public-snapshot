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
)

func TestPermissionConnectionResolver(t *testing.T) {
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

	ps, err := db.Permissions().BulkCreate(ctx, []database.CreatePermissionOpts{
		{
			Namespace: "TEST-NAMESPACE",
			Action:    "READ",
		},
		{
			Namespace: "TEST-NAMESPACE",
			Action:    "WRITE",
		},
		{
			Namespace: "TEST-NAMESPACE",
			Action:    "EXECUTE",
		},
	})
	assert.NoError(t, err)

	want := []apitest.Permission{
		{
			ID: string(marshalPermissionID(ps[0].ID)),
		},
		{
			ID: string(marshalPermissionID(ps[1].ID)),
		},
		{
			ID: string(marshalPermissionID(ps[2].ID)),
		},
	}

	tests := []struct {
		firstParam      int
		wantHasNextPage bool
		wantTotalCount  int
		wantNodes       []apitest.Permission
	}{
		{firstParam: 1, wantHasNextPage: true, wantTotalCount: 3, wantNodes: want[:1]},
		{firstParam: 2, wantHasNextPage: true, wantTotalCount: 3, wantNodes: want[:2]},
		{firstParam: 3, wantHasNextPage: false, wantTotalCount: 3, wantNodes: want},
		{firstParam: 4, wantHasNextPage: false, wantTotalCount: 3, wantNodes: want},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("first=%d", tc.firstParam), func(t *testing.T) {
			input := map[string]any{"first": int64(tc.firstParam)}
			var response struct{ Permissions apitest.PermissionConnection }
			apitest.MustExec(actor.WithActor(context.Background(), actor.FromUser(userID)), t, s, input, &response, queryPermissionConnection)

			wantConnection := apitest.PermissionConnection{
				TotalCount: tc.wantTotalCount,
				PageInfo: apitest.PageInfo{
					HasNextPage: tc.wantHasNextPage,
					// We don't test on the cursor here.
					EndCursor: response.Permissions.PageInfo.EndCursor,
				},
				Nodes: tc.wantNodes,
			}

			if diff := cmp.Diff(wantConnection, response.Permissions); diff != "" {
				t.Fatalf("wrong permssions response (-want +got):\n%s", diff)
			}
		})
	}
}

const queryPermissionConnection = `
query($first: Int, $after: String) {
	permissions(first: $first, after: $after) {
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
	assert.NoError(t, err)

	// create a new role
	role, err := db.Roles().Create(ctx, "TEST-ROLE", false)
	assert.NoError(t, err)

	_, err = db.UserRoles().Create(ctx, database.CreateUserRoleOpts{
		RoleID: role.ID,
		UserID: userID,
	})
	assert.NoError(t, err)

	p, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: "TEST-NAMESPACE",
		Action:    "READ",
	})
	assert.NoError(t, err)

	_, err = db.RolePermissions().Create(ctx, database.CreateRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: p.ID,
	})
	assert.NoError(t, err)

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
