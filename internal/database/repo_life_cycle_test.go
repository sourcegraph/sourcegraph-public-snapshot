package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/stretchr/testify/require"
)

func TestRepoLifeCycle(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := actor.WithInternalActor(context.Background())

	store := db.RepoLifeCycleStore()
	require.NoError(t, store.Upsert(ctx, 1, RepoLifeCycleEventAddedFromCodeHostSync))

	require.NoError(t, store.Upsert(ctx, 1, RepoLifeCycleEventCloneCompleted))

	item, err := store.Get(ctx, 1)
	require.NoError(t, err)

	fmt.Printf("%#v", *item)

}
