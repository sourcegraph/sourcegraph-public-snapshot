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

func TestPermissionResolver(t *testing.T) {
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

	s, err := newSchema(db, &Resolver{
		db:     db,
		logger: logger,
	})
	if err != nil {
		t.Fatal(err)
	}

	mpid := string(marshalPermissionID(perm.ID))

	want := apitest.Permission{
		Typename:  "Permission",
		ID:        mpid,
		Namespace: perm.Namespace,
		Action:    perm.Action,
		CreatedAt: gqlutil.DateTime{Time: perm.CreatedAt.Truncate(time.Second)},
	}

	input := map[string]any{"permission": mpid}
	{
		var response struct{ Node apitest.Permission }
		apitest.MustExec(ctx, t, s, input, &response, queryPermissionNode)
		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	}
}

const queryPermissionNode = `
query ($permission: ID!) {
	node(id: $permission) {
		__typename
		
		... on Permission {
			id
			namespace
			action
			createdAt
		}
	}
}
`
