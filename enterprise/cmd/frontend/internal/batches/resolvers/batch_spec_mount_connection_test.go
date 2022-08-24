package resolvers

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestBatchSpecMountConnectionResolver_TotalCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observation.TestContext, nil)

	specID, err := createBatchSpec(t, db, ctx, bstore)
	require.NoError(t, err)
	err = createBatchSpecMounts(ctx, bstore, specID, 1)
	require.NoError(t, err)

	resolver := batchSpecMountConnectionResolver{
		store: bstore,
		opts: store.ListBatchSpecMountsOpts{
			BatchSpecID: specID,
		},
	}

	count, err := resolver.TotalCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), count)
}

func TestBatchSpecMountConnectionResolver_PageInfo_SinglePage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observation.TestContext, nil)

	specID, err := createBatchSpec(t, db, ctx, bstore)
	require.NoError(t, err)
	err = createBatchSpecMounts(ctx, bstore, specID, 1)
	require.NoError(t, err)

	resolver := batchSpecMountConnectionResolver{
		store: bstore,
		opts: store.ListBatchSpecMountsOpts{
			BatchSpecID: specID,
		},
	}

	pageInfo, err := resolver.PageInfo(ctx)
	assert.NoError(t, err)
	assert.False(t, pageInfo.HasNextPage())
	assert.Nil(t, pageInfo.EndCursor())
}

func TestBatchSpecMountConnectionResolver_PageInfo_MultiplePages(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	bstore := store.New(db, &observation.TestContext, nil)

	specID, err := createBatchSpec(t, db, ctx, bstore)
	require.NoError(t, err)
	err = createBatchSpecMounts(ctx, bstore, specID, 10)
	require.NoError(t, err)

	resolver := batchSpecMountConnectionResolver{
		store: bstore,
		opts: store.ListBatchSpecMountsOpts{
			LimitOpts: store.LimitOpts{
				Limit: 5,
			},
			BatchSpecID: specID,
		},
	}

	pageInfo, err := resolver.PageInfo(ctx)
	assert.NoError(t, err)
	assert.True(t, pageInfo.HasNextPage())
	assert.NotNil(t, pageInfo.EndCursor())

	cursor, err := strconv.ParseInt(*pageInfo.EndCursor(), 10, 32)
	require.NoError(t, err)
	resolver = batchSpecMountConnectionResolver{
		store: bstore,
		opts: store.ListBatchSpecMountsOpts{
			LimitOpts: store.LimitOpts{
				Limit: 5,
			},
			BatchSpecID: specID,
			Cursor:      cursor,
		},
	}

	pageInfo, err = resolver.PageInfo(ctx)
	assert.NoError(t, err)
	assert.False(t, pageInfo.HasNextPage())
	assert.Nil(t, pageInfo.EndCursor())
}

func TestBatchSpecMountConnectionResolver_Nodes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observation.TestContext, nil)

	specID, err := createBatchSpec(t, db, ctx, bstore)
	require.NoError(t, err)
	err = createBatchSpecMounts(ctx, bstore, specID, 1)
	require.NoError(t, err)

	resolver := batchSpecMountConnectionResolver{
		store: bstore,
		opts: store.ListBatchSpecMountsOpts{
			BatchSpecID: specID,
		},
	}

	nodes, err := resolver.Nodes(ctx)
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
}

func TestBatchSpecMountConnectionResolver_Nodes_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observation.TestContext, nil)

	specID, err := createBatchSpec(t, db, ctx, bstore)
	require.NoError(t, err)

	resolver := batchSpecMountConnectionResolver{
		store: bstore,
		opts: store.ListBatchSpecMountsOpts{
			BatchSpecID: specID,
		},
	}

	nodes, err := resolver.Nodes(ctx)
	assert.NoError(t, err)
	assert.Len(t, nodes, 0)
}

func createBatchSpec(t *testing.T, db database.DB, ctx context.Context, bstore *store.Store) (int64, error) {
	userID := bt.CreateTestUser(t, db, true).ID
	spec := &btypes.BatchSpec{
		NamespaceUserID: userID,
		UserID:          userID,
	}
	if err := bstore.CreateBatchSpec(ctx, spec); err != nil {
		return 0, err
	}
	return spec.ID, nil
}

func createBatchSpecMounts(ctx context.Context, bstore *store.Store, specID int64, count int) error {
	for i := 0; i < count; i++ {
		mount := &btypes.BatchSpecMount{
			BatchSpecID: specID,
			FileName:    fmt.Sprintf("hello-%d.txt", i),
			Path:        "foo/bar",
			Size:        12,
			ModifiedAt:  time.Now().UTC(),
		}
		if err := bstore.UpsertBatchSpecMount(ctx, mount); err != nil {
			return err
		}
	}
	return nil
}
