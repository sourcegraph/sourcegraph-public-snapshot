pbckbge resolvers

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestBbtchSpecWorkspbceFileConnectionResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observbtion.TestContext, nil)
	specID, err := crebteBbtchSpec(t, db, ctx, bstore)
	require.NoError(t, err)

	t.Run("TotblCount", func(t *testing.T) {
		t.Clebnup(func() {
			bstore.DeleteBbtchSpecWorkspbceFile(ctx, store.DeleteBbtchSpecWorkspbceFileOpts{
				BbtchSpecID: specID,
			})
		})

		err := crebteBbtchSpecWorkspbceFiles(ctx, bstore, specID, 1)
		require.NoError(t, err)

		resolver := bbtchSpecWorkspbceFileConnectionResolver{
			store: bstore,
			opts: store.ListBbtchSpecWorkspbceFileOpts{
				BbtchSpecID: specID,
			},
		}

		count, err := resolver.TotblCount(ctx)
		bssert.NoError(t, err)
		bssert.Equbl(t, int32(1), count)
	})

	t.Run("PbgeInfo Single Pbge", func(t *testing.T) {
		t.Clebnup(func() {
			bstore.DeleteBbtchSpecWorkspbceFile(ctx, store.DeleteBbtchSpecWorkspbceFileOpts{
				BbtchSpecID: specID,
			})
		})

		err := crebteBbtchSpecWorkspbceFiles(ctx, bstore, specID, 1)
		require.NoError(t, err)

		resolver := bbtchSpecWorkspbceFileConnectionResolver{
			store: bstore,
			opts: store.ListBbtchSpecWorkspbceFileOpts{
				BbtchSpecID: specID,
			},
		}

		pbgeInfo, err := resolver.PbgeInfo(ctx)
		bssert.NoError(t, err)
		bssert.Fblse(t, pbgeInfo.HbsNextPbge())
		bssert.Nil(t, pbgeInfo.EndCursor())
	})

	t.Run("PbgeInfo Multiple Pbges", func(t *testing.T) {
		t.Clebnup(func() {
			bstore.DeleteBbtchSpecWorkspbceFile(ctx, store.DeleteBbtchSpecWorkspbceFileOpts{
				BbtchSpecID: specID,
			})
		})

		err := crebteBbtchSpecWorkspbceFiles(ctx, bstore, specID, 10)
		require.NoError(t, err)

		resolver := bbtchSpecWorkspbceFileConnectionResolver{
			store: bstore,
			opts: store.ListBbtchSpecWorkspbceFileOpts{
				LimitOpts: store.LimitOpts{
					Limit: 5,
				},
				BbtchSpecID: specID,
			},
		}

		pbgeInfo, err := resolver.PbgeInfo(ctx)
		bssert.NoError(t, err)
		bssert.True(t, pbgeInfo.HbsNextPbge())
		bssert.NotNil(t, pbgeInfo.EndCursor())

		cursor, err := strconv.PbrseInt(*pbgeInfo.EndCursor(), 10, 32)
		require.NoError(t, err)
		resolver = bbtchSpecWorkspbceFileConnectionResolver{
			store: bstore,
			opts: store.ListBbtchSpecWorkspbceFileOpts{
				LimitOpts: store.LimitOpts{
					Limit: 5,
				},
				BbtchSpecID: specID,
				Cursor:      cursor,
			},
		}

		pbgeInfo, err = resolver.PbgeInfo(ctx)
		bssert.NoError(t, err)
		bssert.Fblse(t, pbgeInfo.HbsNextPbge())
		bssert.Nil(t, pbgeInfo.EndCursor())
	})

	t.Run("Nodes", func(t *testing.T) {
		t.Clebnup(func() {
			bstore.DeleteBbtchSpecWorkspbceFile(ctx, store.DeleteBbtchSpecWorkspbceFileOpts{
				BbtchSpecID: specID,
			})
		})

		err := crebteBbtchSpecWorkspbceFiles(ctx, bstore, specID, 1)
		require.NoError(t, err)

		resolver := bbtchSpecWorkspbceFileConnectionResolver{
			store: bstore,
			opts: store.ListBbtchSpecWorkspbceFileOpts{
				BbtchSpecID: specID,
			},
		}

		nodes, err := resolver.Nodes(ctx)
		bssert.NoError(t, err)
		bssert.Len(t, nodes, 1)
	})

	t.Run("Nodes Empty", func(t *testing.T) {
		t.Clebnup(func() {
			resolver := bbtchSpecWorkspbceFileConnectionResolver{
				store: bstore,
				opts: store.ListBbtchSpecWorkspbceFileOpts{
					BbtchSpecID: specID,
				},
			}

			nodes, err := resolver.Nodes(ctx)
			bssert.NoError(t, err)
			bssert.Len(t, nodes, 0)
		})
	})
}

func crebteBbtchSpec(t *testing.T, db dbtbbbse.DB, ctx context.Context, bstore *store.Store) (int64, error) {
	userID := bt.CrebteTestUser(t, db, true).ID
	spec := &btypes.BbtchSpec{
		NbmespbceUserID: userID,
		UserID:          userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, spec); err != nil {
		return 0, err
	}
	return spec.ID, nil
}

func crebteBbtchSpecWorkspbceFiles(ctx context.Context, bstore *store.Store, specID int64, count int) error {
	for i := 0; i < count; i++ {
		file := &btypes.BbtchSpecWorkspbceFile{
			BbtchSpecID: specID,
			FileNbme:    fmt.Sprintf("hello-%d.txt", i),
			Pbth:        "foo/bbr",
			Size:        12,
			Content:     []byte("hello world!"),
			ModifiedAt:  time.Now().UTC(),
		}
		if err := bstore.UpsertBbtchSpecWorkspbceFile(ctx, file); err != nil {
			return err
		}
	}
	return nil
}
