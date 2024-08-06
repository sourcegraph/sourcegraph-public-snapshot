package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
)

func TestPermissionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)

	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))

	user := createTestUser(t, db, false)
	admin := createTestUser(t, db, true)

	userCtx := actor.WithActor(ctx, actor.FromUser(user.ID))
	adminCtx := actor.WithActor(ctx, actor.FromUser(admin.ID))

	perm, err := db.Permissions().Create(ctx, database.CreatePermissionOpts{
		Namespace: rtypes.BatchChangesNamespace,
		Action:    rtypes.BatchChangesReadAction,
	})
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSchemaWithoutResolvers(db, nil)
	if err != nil {
		t.Fatal(err)
	}

	mpid := string(MarshalPermissionID(perm.ID))

	t.Run("as non site-administrator", func(t *testing.T) {
		input := map[string]any{"permission": mpid}
		var response struct{ Node apitest.Permission }
		errs := apitest.Exec(userCtx, t, s, input, &response, queryPermissionNode)

		require.Len(t, errs, 1)
		require.Equal(t, errs[0].Message, "must be site admin")
	})

	t.Run("as site-administrator", func(t *testing.T) {
		want := apitest.Permission{
			Typename:    "Permission",
			ID:          mpid,
			Namespace:   perm.Namespace,
			DisplayName: perm.DisplayName(),
			Action:      perm.Action,
			CreatedAt:   gqlutil.DateTime{Time: perm.CreatedAt.Truncate(time.Second)},
		}

		input := map[string]any{"permission": mpid}
		var response struct{ Node apitest.Permission }
		apitest.MustExec(adminCtx, t, s, input, &response, queryPermissionNode)
		if diff := cmp.Diff(want, response.Node); diff != "" {
			t.Fatalf("unexpected response (-want +got):\n%s", diff)
		}
	})
}

const queryPermissionNode = `
query ($permission: ID!) {
	node(id: $permission) {
		__typename

		... on Permission {
			id
			namespace
			displayName
			action
			createdAt
		}
	}
}
`
