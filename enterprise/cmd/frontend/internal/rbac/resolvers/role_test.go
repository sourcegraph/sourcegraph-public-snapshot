package resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/rbac/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func TestRoleResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	perm, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: "BATCHCHANGES",
		Action:    "READ",
	})
	if err != nil {
		t.Fatal(err)
	}

	role, err := db.Roles().Create(ctx, "BATCHCHANGES_ADMIN", false)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.RolePermissions().Create(ctx, database.CreateRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: perm.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	s, err := newSchema(db, &Resolver{
		db:     db,
		logger: logger,
	})
	if err != nil {
		t.Fatal(err)
	}

	mrid := string(marshalRoleID(role.ID))
	mpid := string(marshalPermissionID(perm.ID))

	want := apitest.Role{
		Typename:  "Role",
		ID:        mrid,
		Name:      role.Name,
		Readonly:  role.System,
		CreatedAt: gqlutil.DateTime{Time: role.CreatedAt.Truncate(time.Second)},
		DeletedAt: nil,
		Permissions: apitest.PermissionConnection{
			TotalCount: 1,
			PageInfo: apitest.PageInfo{
				HasNextPage: false,
				EndCursor:   nil,
			},
			Nodes: []apitest.Permission{
				{
					ID:        mpid,
					Namespace: perm.Namespace,
					Action:    perm.Action,
					CreatedAt: gqlutil.DateTime{Time: perm.CreatedAt.Truncate(time.Second)},
				},
			},
		},
	}

	input := map[string]any{"role": mrid}
	{
		var response struct{ Node apitest.Permission }
		apitest.MustExec(ctx, t, s, input, &response, queryRoleNode)
		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	}
}

const queryRoleNode = `
query ($role: ID!) {
	node(id: $role) {
		__typename

		... on Role {
			id
			name
			readonly
			createdAt
			permissions {
				nodes {
					id
					namespace
					action
					createdAt
				}
				totalCount
				pageInfo {
					endCursor
					hasNextPage
				}
			}
		}
	}
}
`
