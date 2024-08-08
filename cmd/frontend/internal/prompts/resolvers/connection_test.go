package resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestPromptsConnectionStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(logtest.Scoped(t), dbtest.NewDB(t))

	user, err := db.Users().Create(ctx, database.NewUser{
		Email:           "test@sourcegraph.com",
		Username:        "test",
		EmailIsVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, &actor.Actor{UID: user.ID})

	for i := range 10 {
		created, err := db.Prompts().Create(ctx, &types.Prompt{
			Name:  fmt.Sprintf("prompt-%d", i),
			Owner: types.NamespaceUser(user.ID),
		})
		if err != nil {
			t.Fatal(err)
		}

		// Adjust so each one has a different updated_at value (which is rounded to the second).
		if _, err := db.ExecContext(ctx, `UPDATE prompts SET created_at = '2024-07-04 12:34:56.123456', updated_at = '2024-07-05 19:46:03.515814'::timestamp with time zone - (INTERVAL '100 milliseconds' * $1) WHERE id = $2`, i, created.ID); err != nil {
			t.Fatal(err)
		}
	}

	owner := types.NamespaceUser(user.ID)
	connectionStore := &promptsConnectionStore{
		db:       db,
		listArgs: database.PromptListArgs{Owner: &owner},
	}

	t.Run("no orderBy", func(t *testing.T) {
		gqlutil.TestConnectionResolverStoreSuite(t, connectionStore, nil)
	})

	t.Run("orderBy updated_at", func(t *testing.T) {
		var pgArgs gqlutil.TestPaginationArgs
		pgArgs.OrderBy, pgArgs.Ascending = database.PromptsOrderByUpdatedAt.ToOptions()
		gqlutil.TestConnectionResolverStoreSuite(t, connectionStore, &pgArgs)
	})

	t.Run("orderBy name_with_owner", func(t *testing.T) {
		var pgArgs gqlutil.TestPaginationArgs
		pgArgs.OrderBy, pgArgs.Ascending = database.PromptsOrderByNameWithOwner.ToOptions()
		gqlutil.TestConnectionResolverStoreSuite(t, connectionStore, &pgArgs)
	})
}

var dummyConnectionResolverArgs = gqlutil.ConnectionResolverArgs{First: pointers.Ptr[int32](1)}
