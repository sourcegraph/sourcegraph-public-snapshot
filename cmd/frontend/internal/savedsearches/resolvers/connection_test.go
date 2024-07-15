package resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestSavedSearchesConnectionStore(t *testing.T) {
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

	for i := range 10 {
		created, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
			Description: fmt.Sprintf("Test Search %d", i),
			Query:       "r:src-cli",
			Owner:       types.NamespaceUser(user.ID),
		})
		if err != nil {
			t.Fatal(err)
		}

		// Adjust so each one has a different updated_at value (which is rounded to the second).
		if _, err := db.ExecContext(ctx, `UPDATE saved_searches SET created_at = '2024-07-04 12:34:56.123456', updated_at = '2024-07-05 19:46:03.515814'::timestamp with time zone - (INTERVAL '100 milliseconds' * $1) WHERE id = $2`, i, created.ID); err != nil {
			t.Fatal(err)
		}
	}

	owner := types.NamespaceUser(user.ID)
	connectionStore := &savedSearchesConnectionStore{
		db:       db,
		listArgs: database.SavedSearchListArgs{Owner: &owner},
	}

	t.Run("no orderBy", func(t *testing.T) {
		graphqlutil.TestConnectionResolverStoreSuite(t, connectionStore, nil)
	})

	t.Run("orderBy updated_at", func(t *testing.T) {
		var pgArgs graphqlutil.TestPaginationArgs
		pgArgs.OrderBy, pgArgs.Ascending = database.SavedSearchesOrderByUpdatedAt.ToOptions()
		graphqlutil.TestConnectionResolverStoreSuite(t, connectionStore, &pgArgs)
	})

	t.Run("orderBy description", func(t *testing.T) {
		var pgArgs graphqlutil.TestPaginationArgs
		pgArgs.OrderBy, pgArgs.Ascending = database.SavedSearchesOrderByDescription.ToOptions()
		graphqlutil.TestConnectionResolverStoreSuite(t, connectionStore, &pgArgs)
	})
}

var dummyConnectionResolverArgs = graphqlutil.ConnectionResolverArgs{First: pointers.Ptr[int32](1)}
