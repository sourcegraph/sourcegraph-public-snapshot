package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
)

func TestRoleResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))

	userID := createTestUser(t, db, false).ID
	userCtx := actor.WithActor(ctx, actor.FromUser(userID))

	adminUserID := createTestUser(t, db, true).ID
	adminCtx := actor.WithActor(ctx, actor.FromUser(adminUserID))

	perm, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: rtypes.BatchChangesNamespace,
		Action:    "READ",
	})
	if err != nil {
		t.Fatal(err)
	}

	role, err := db.Roles().Create(ctx, "BATCHCHANGES_ADMIN", false)
	if err != nil {
		t.Fatal(err)
	}

	err = db.RolePermissions().Assign(ctx, database.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perm.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSchemaWithoutResolvers(db)
	if err != nil {
		t.Fatal(err)
	}

	mrid := string(MarshalRoleID(role.ID))
	mpid := string(MarshalPermissionID(perm.ID))

	t.Run("as site-administrator", func(t *testing.T) {
		want := apitest.Role{
			Typename:  "Role",
			ID:        mrid,
			Name:      role.Name,
			System:    role.System,
			CreatedAt: gqlutil.DateTime{Time: role.CreatedAt.Truncate(time.Second)},
			DeletedAt: nil,
			Permissions: apitest.PermissionConnection{
				TotalCount: 1,
				PageInfo: apitest.PageInfo{
					HasNextPage:     false,
					HasPreviousPage: false,
				},
				Nodes: []apitest.Permission{
					{
						ID:          mpid,
						Namespace:   perm.Namespace,
						DisplayName: perm.DisplayName(),
						Action:      perm.Action,
						CreatedAt:   gqlutil.DateTime{Time: perm.CreatedAt.Truncate(time.Second)},
					},
				},
			},
		}

		input := map[string]any{"role": mrid}
		var response struct{ Node apitest.Role }
		apitest.MustExec(adminCtx, t, s, input, &response, queryRoleNode)
		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})

	t.Run("non site-administrator", func(t *testing.T) {
		input := map[string]any{"role": mrid}
		var response struct{ Node apitest.Role }
		errs := apitest.Exec(userCtx, t, s, input, &response, queryRoleNode)

		assert.Len(t, errs, 1)
		assert.Equal(t, errs[0].Message, "must be site admin")
	})
}

const queryRoleNode = `
query ($role: ID!) {
	node(id: $role) {
		__typename

		... on Role {
			id
			name
			system
			createdAt
			permissions(first: 50) {
				nodes {
					id
					namespace
					displayName
					action
					createdAt
				}
				totalCount
				pageInfo {
					hasPreviousPage
					hasNextPage
				}
			}
		}
	}
}
`
