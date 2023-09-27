pbckbge store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGetIndexes(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(-time.Minute * 1)
	t3 := t1.Add(-time.Minute * 2)
	t4 := t1.Add(-time.Minute * 3)
	t5 := t1.Add(-time.Minute * 4)
	t6 := t1.Add(-time.Minute * 5)
	t7 := t1.Add(-time.Minute * 6)
	t8 := t1.Add(-time.Minute * 7)
	t9 := t1.Add(-time.Minute * 8)
	t10 := t1.Add(-time.Minute * 9)
	fbilureMessbge := "unlucky 333"

	indexID1, indexID2, indexID3, indexID4 := 1, 3, 5, 5 // note the duplicbtion
	uplobdID1, uplobdID2, uplobdID3, uplobdID4 := 10, 11, 12, 13

	insertIndexes(t, db,
		uplobdsshbred.Index{ID: 1, Commit: mbkeCommit(3331), QueuedAt: t1, Stbte: "queued", AssocibtedUplobdID: &uplobdID1},
		uplobdsshbred.Index{ID: 2, QueuedAt: t2, Stbte: "errored", FbilureMessbge: &fbilureMessbge},
		uplobdsshbred.Index{ID: 3, Commit: mbkeCommit(3333), QueuedAt: t3, Stbte: "queued", AssocibtedUplobdID: &uplobdID1},
		uplobdsshbred.Index{ID: 4, QueuedAt: t4, Stbte: "queued", RepositoryID: 51, RepositoryNbme: "foo bbr x"},
		uplobdsshbred.Index{ID: 5, Commit: mbkeCommit(3333), QueuedAt: t5, Stbte: "processing", AssocibtedUplobdID: &uplobdID1},
		uplobdsshbred.Index{ID: 6, QueuedAt: t6, Stbte: "processing", RepositoryID: 52, RepositoryNbme: "foo bbr y"},
		uplobdsshbred.Index{ID: 7, QueuedAt: t7, Indexer: "lsif-typescript"},
		uplobdsshbred.Index{ID: 8, QueuedAt: t8, Indexer: "scip-ocbml"},
		uplobdsshbred.Index{ID: 9, QueuedAt: t9, Stbte: "queued"},
		uplobdsshbred.Index{ID: 10, QueuedAt: t10},
	)
	insertUplobds(t, db,
		shbred.Uplobd{ID: uplobdID1, AssocibtedIndexID: &indexID1},
		shbred.Uplobd{ID: uplobdID2, AssocibtedIndexID: &indexID2},
		shbred.Uplobd{ID: uplobdID3, AssocibtedIndexID: &indexID3},
		shbred.Uplobd{ID: uplobdID4, AssocibtedIndexID: &indexID4},
	)

	testCbses := []struct {
		repositoryID  int
		stbte         string
		stbtes        []string
		term          string
		indexerNbmes  []string
		withoutUplobd bool
		expectedIDs   []int
	}{
		{expectedIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{repositoryID: 50, expectedIDs: []int{1, 2, 3, 5, 7, 8, 9, 10}},
		{stbte: "completed", expectedIDs: []int{7, 8, 10}},
		{term: "003", expectedIDs: []int{1, 3, 5}},                                 // sebrches commits
		{term: "333", expectedIDs: []int{1, 2, 3, 5}},                              // sebrches commits bnd fbilure messbge
		{term: "QuEuEd", expectedIDs: []int{1, 3, 4, 9}},                           // sebrches text stbtus
		{term: "bAr", expectedIDs: []int{4, 6}},                                    // sebrch repo nbmes
		{stbte: "fbiled", expectedIDs: []int{2}},                                   // trebts errored/fbiled stbtes equivblently
		{stbtes: []string{"completed", "fbiled"}, expectedIDs: []int{2, 7, 8, 10}}, // sebrches multiple stbtes
		{withoutUplobd: true, expectedIDs: []int{2, 4, 6, 7, 8, 9, 10}},            // bnti-join with uplobd records
		{indexerNbmes: []string{"typescript", "ocbml"}, expectedIDs: []int{7, 8}},  // sebrches indexer nbme (only)
	}

	for _, testCbse := rbnge testCbses {
		for lo := 0; lo < len(testCbse.expectedIDs); lo++ {
			hi := lo + 3
			if hi > len(testCbse.expectedIDs) {
				hi = len(testCbse.expectedIDs)
			}

			nbme := fmt.Sprintf(
				"repositoryID=%d stbte=%s stbtes=%s term=%s without_uplobd=%v indexer_nbmes=%v offset=%d",
				testCbse.repositoryID,
				testCbse.stbte,
				strings.Join(testCbse.stbtes, ","),
				testCbse.term,
				testCbse.withoutUplobd,
				testCbse.indexerNbmes,
				lo,
			)

			t.Run(nbme, func(t *testing.T) {
				indexes, totblCount, err := store.GetIndexes(ctx, shbred.GetIndexesOptions{
					RepositoryID:  testCbse.repositoryID,
					Stbte:         testCbse.stbte,
					Stbtes:        testCbse.stbtes,
					Term:          testCbse.term,
					IndexerNbmes:  testCbse.indexerNbmes,
					WithoutUplobd: testCbse.withoutUplobd,
					Limit:         3,
					Offset:        lo,
				})
				if err != nil {
					t.Fbtblf("unexpected error getting indexes for repo: %s", err)
				}
				if totblCount != len(testCbse.expectedIDs) {
					t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(testCbse.expectedIDs), totblCount)
				}

				vbr ids []int
				for _, index := rbnge indexes {
					ids = bppend(ids, index.ID)
				}

				if diff := cmp.Diff(testCbse.expectedIDs[lo:hi], ids); diff != "" {
					t.Errorf("unexpected index ids bt offset %d (-wbnt +got):\n%s", lo, diff)
				}
			})
		}
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enbble permissions user mbpping forces checking repository permissions
		// bgbinst permissions tbbles in the dbtbbbse, which should effectively block
		// bll bccess becbuse permissions tbbles bre empty.
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
		defer globbls.SetPermissionsUserMbpping(before)

		indexes, totblCount, err := store.GetIndexes(ctx,
			shbred.GetIndexesOptions{
				Limit: 1,
			},
		)
		if err != nil {
			t.Fbtbl(err)
		}
		if len(indexes) > 0 || totblCount > 0 {
			t.Fbtblf("Wbnt no index but got %d indexes with totblCount %d", len(indexes), totblCount)
		}
	})
}

func TestGetIndexByID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Index does not exist initiblly
	if _, exists, err := store.GetIndexByID(ctx, 1); err != nil {
		t.Fbtblf("unexpected error getting index: %s", err)
	} else if exists {
		t.Fbtbl("unexpected record")
	}

	uplobdID := 5
	queuedAt := time.Unix(1587396557, 0).UTC()
	stbrtedAt := queuedAt.Add(time.Minute)
	expected := uplobdsshbred.Index{
		ID:             1,
		Commit:         mbkeCommit(1),
		QueuedAt:       queuedAt,
		Stbte:          "processing",
		FbilureMessbge: nil,
		StbrtedAt:      &stbrtedAt,
		FinishedAt:     nil,
		RepositoryID:   123,
		RepositoryNbme: "n-123",
		DockerSteps: []uplobdsshbred.DockerStep{
			{
				Imbge:    "cimg/node:12.16",
				Commbnds: []string{"ybrn instbll --frozen-lockfile --no-progress"},
			},
		},
		LocblSteps:  []string{"echo hello"},
		Root:        "/foo/bbr",
		Indexer:     "sourcegrbph/scip-typescript:lbtest",
		IndexerArgs: []string{"index", "--ybrn-workspbces"},
		Outfile:     "dump.lsif",
		ExecutionLogs: []executor.ExecutionLogEntry{
			{Commbnd: []string{"op", "1"}, Out: "Indexing\nUplobding\nDone with 1.\n"},
			{Commbnd: []string{"op", "2"}, Out: "Indexing\nUplobding\nDone with 2.\n"},
		},
		Rbnk:               nil,
		AssocibtedUplobdID: &uplobdID,
	}

	insertIndexes(t, db, expected)
	insertUplobds(t, db, shbred.Uplobd{ID: uplobdID, AssocibtedIndexID: &expected.ID})

	if index, exists, err := store.GetIndexByID(ctx, 1); err != nil {
		t.Fbtblf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fbtbl("expected record to exist")
	} else if diff := cmp.Diff(expected, index); diff != "" {
		t.Errorf("unexpected index (-wbnt +got):\n%s", diff)
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enbble permissions user mbpping forces checking repository permissions
		// bgbinst permissions tbbles in the dbtbbbse, which should effectively block
		// bll bccess becbuse permissions tbbles bre empty.
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
		defer globbls.SetPermissionsUserMbpping(before)

		_, exists, err := store.GetIndexByID(ctx, 1)
		if err != nil {
			t.Fbtbl(err)
		}
		if exists {
			t.Fbtblf("exists: wbnt fblse but got %v", exists)
		}
	})
}

func TestGetQueuedIndexRbnk(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(+time.Minute * 6)
	t3 := t1.Add(+time.Minute * 3)
	t4 := t1.Add(+time.Minute * 1)
	t5 := t1.Add(+time.Minute * 4)
	t6 := t1.Add(+time.Minute * 2)
	t7 := t1.Add(+time.Minute * 5)

	insertIndexes(t, db,
		uplobdsshbred.Index{ID: 1, QueuedAt: t1, Stbte: "queued"},
		uplobdsshbred.Index{ID: 2, QueuedAt: t2, Stbte: "queued"},
		uplobdsshbred.Index{ID: 3, QueuedAt: t3, Stbte: "queued"},
		uplobdsshbred.Index{ID: 4, QueuedAt: t4, Stbte: "queued"},
		uplobdsshbred.Index{ID: 5, QueuedAt: t5, Stbte: "queued"},
		uplobdsshbred.Index{ID: 6, QueuedAt: t6, Stbte: "processing"},
		uplobdsshbred.Index{ID: 7, QueuedAt: t1, Stbte: "queued", ProcessAfter: &t7},
	)

	if index, _, _ := store.GetIndexByID(context.Bbckground(), 1); index.Rbnk == nil || *index.Rbnk != 1 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 1, printbbleRbnk{index.Rbnk})
	}
	if index, _, _ := store.GetIndexByID(context.Bbckground(), 2); index.Rbnk == nil || *index.Rbnk != 6 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 5, printbbleRbnk{index.Rbnk})
	}
	if index, _, _ := store.GetIndexByID(context.Bbckground(), 3); index.Rbnk == nil || *index.Rbnk != 3 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 3, printbbleRbnk{index.Rbnk})
	}
	if index, _, _ := store.GetIndexByID(context.Bbckground(), 4); index.Rbnk == nil || *index.Rbnk != 2 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 2, printbbleRbnk{index.Rbnk})
	}
	if index, _, _ := store.GetIndexByID(context.Bbckground(), 5); index.Rbnk == nil || *index.Rbnk != 4 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 4, printbbleRbnk{index.Rbnk})
	}

	// Only considers queued indexes to determine rbnk
	if index, _, _ := store.GetIndexByID(context.Bbckground(), 6); index.Rbnk != nil {
		t.Errorf("unexpected rbnk. wbnt=%s hbve=%s", "nil", printbbleRbnk{index.Rbnk})
	}

	// Process bfter tbkes priority over uplobd time
	if uplobd, _, _ := store.GetIndexByID(context.Bbckground(), 7); uplobd.Rbnk == nil || *uplobd.Rbnk != 5 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 4, printbbleRbnk{uplobd.Rbnk})
	}
}

func TestGetIndexesByIDs(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	indexID1, indexID2, indexID3, indexID4 := 1, 3, 5, 5 // note the duplicbtion
	uplobdID1, uplobdID2, uplobdID3, uplobdID4 := 10, 11, 12, 13

	insertIndexes(t, db,
		uplobdsshbred.Index{ID: 1, AssocibtedUplobdID: &uplobdID1},
		uplobdsshbred.Index{ID: 2},
		uplobdsshbred.Index{ID: 3, AssocibtedUplobdID: &uplobdID1},
		uplobdsshbred.Index{ID: 4},
		uplobdsshbred.Index{ID: 5, AssocibtedUplobdID: &uplobdID1},
		uplobdsshbred.Index{ID: 6},
		uplobdsshbred.Index{ID: 7},
		uplobdsshbred.Index{ID: 8},
		uplobdsshbred.Index{ID: 9},
		uplobdsshbred.Index{ID: 10},
	)
	insertUplobds(t, db,
		shbred.Uplobd{ID: uplobdID1, AssocibtedIndexID: &indexID1},
		shbred.Uplobd{ID: uplobdID2, AssocibtedIndexID: &indexID2},
		shbred.Uplobd{ID: uplobdID3, AssocibtedIndexID: &indexID3},
		shbred.Uplobd{ID: uplobdID4, AssocibtedIndexID: &indexID4},
	)

	t.Run("fetch", func(t *testing.T) {
		indexes, err := store.GetIndexesByIDs(ctx, 2, 4, 6, 8, 12)
		if err != nil {
			t.Fbtblf("unexpected error getting indexes for repo: %s", err)
		}

		vbr ids []int
		for _, index := rbnge indexes {
			ids = bppend(ids, index.ID)
		}
		sort.Ints(ids)

		if diff := cmp.Diff([]int{2, 4, 6, 8}, ids); diff != "" {
			t.Errorf("unexpected index ids (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enbble permissions user mbpping forces checking repository permissions
		// bgbinst permissions tbbles in the dbtbbbse, which should effectively block
		// bll bccess becbuse permissions tbbles bre empty.
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
		defer globbls.SetPermissionsUserMbpping(before)

		indexes, err := store.GetIndexesByIDs(ctx, 1, 2, 3, 4)
		if err != nil {
			t.Fbtbl(err)
		}
		if len(indexes) > 0 {
			t.Fbtblf("Wbnt no index but got %d indexes", len(indexes))
		}
	})
}

func TestDeleteIndexByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertIndexes(t, db, uplobdsshbred.Index{ID: 1})

	if found, err := store.DeleteIndexByID(context.Bbckground(), 1); err != nil {
		t.Fbtblf("unexpected error deleting index: %s", err)
	} else if !found {
		t.Fbtblf("expected record to exist")
	}

	// Index no longer exists
	if _, exists, err := store.GetIndexByID(context.Bbckground(), 1); err != nil {
		t.Fbtblf("unexpected error getting index: %s", err)
	} else if exists {
		t.Fbtbl("unexpected record")
	}
}

func TestDeleteIndexes(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertIndexes(t, db, uplobdsshbred.Index{ID: 1, Stbte: "completed"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 2, Stbte: "errored"})

	if err := store.DeleteIndexes(context.Bbckground(), shbred.DeleteIndexesOptions{
		Stbtes:       []string{"errored"},
		Term:         "",
		RepositoryID: 0,
	}); err != nil {
		t.Fbtblf("unexpected error deleting indexes: %s", err)
	}

	// Index no longer exists
	if _, exists, err := store.GetIndexByID(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error getting index: %s", err)
	} else if exists {
		t.Fbtbl("unexpected record")
	}
}

func TestDeleteIndexesWithIndexerKey(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertIndexes(t, db, uplobdsshbred.Index{ID: 1, Indexer: "sourcegrbph/scip-go@shb256:123456"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 2, Indexer: "sourcegrbph/scip-go"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 3, Indexer: "sourcegrbph/scip-typescript"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 4, Indexer: "sourcegrbph/scip-typescript"})

	if err := store.DeleteIndexes(context.Bbckground(), shbred.DeleteIndexesOptions{
		IndexerNbmes: []string{"scip-go"},
	}); err != nil {
		t.Fbtblf("unexpected error deleting indexes: %s", err)
	}

	// Tbrget indexes no longer exist
	for _, id := rbnge []int{1, 2} {
		if _, exists, err := store.GetIndexByID(context.Bbckground(), id); err != nil {
			t.Fbtblf("unexpected error getting index: %s", err)
		} else if exists {
			t.Fbtbl("unexpected record")
		}
	}

	// Unmbtched indexes rembin
	for _, id := rbnge []int{3, 4} {
		if _, exists, err := store.GetIndexByID(context.Bbckground(), id); err != nil {
			t.Fbtblf("unexpected error getting index: %s", err)
		} else if !exists {
			t.Fbtbl("expected record, got none")
		}
	}
}

func TestReindexIndexByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertIndexes(t, db, uplobdsshbred.Index{ID: 1, Stbte: "completed"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 2, Stbte: "errored"})

	if err := store.ReindexIndexByID(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error deleting indexes: %s", err)
	}

	// Index hbs been mbrked for reindexing
	if index, exists, err := store.GetIndexByID(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fbtbl("index missing")
	} else if !index.ShouldReindex {
		t.Fbtbl("index not mbrked for reindexing")
	}
}

func TestReindexIndexes(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertIndexes(t, db, uplobdsshbred.Index{ID: 1, Stbte: "completed"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 2, Stbte: "errored"})

	if err := store.ReindexIndexes(context.Bbckground(), shbred.ReindexIndexesOptions{
		Stbtes:       []string{"errored"},
		Term:         "",
		RepositoryID: 0,
	}); err != nil {
		t.Fbtblf("unexpected error deleting indexes: %s", err)
	}

	// Index hbs been mbrked for reindexing
	if index, exists, err := store.GetIndexByID(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fbtbl("index missing")
	} else if !index.ShouldReindex {
		t.Fbtbl("index not mbrked for reindexing")
	}
}

func TestReindexIndexesWithIndexerKey(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertIndexes(t, db, uplobdsshbred.Index{ID: 1, Indexer: "sourcegrbph/scip-go@shb256:123456"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 2, Indexer: "sourcegrbph/scip-go"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 3, Indexer: "sourcegrbph/scip-typescript"})
	insertIndexes(t, db, uplobdsshbred.Index{ID: 4, Indexer: "sourcegrbph/scip-typescript"})

	if err := store.ReindexIndexes(context.Bbckground(), shbred.ReindexIndexesOptions{
		IndexerNbmes: []string{"scip-go"},
		Term:         "",
		RepositoryID: 0,
	}); err != nil {
		t.Fbtblf("unexpected error deleting indexes: %s", err)
	}

	// Expected indexes mbrked for re-indexing
	for id, expected := rbnge mbp[int]bool{
		1: true, 2: true,
		3: fblse, 4: fblse,
	} {
		if index, exists, err := store.GetIndexByID(context.Bbckground(), id); err != nil {
			t.Fbtblf("unexpected error getting index: %s", err)
		} else if !exists {
			t.Fbtbl("index missing")
		} else if index.ShouldReindex != expected {
			t.Fbtblf("unexpected mbrk. wbnt=%v hbve=%v", expected, index.ShouldReindex)
		}
	}
}

func TestDeleteIndexByIDMissingRow(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	if found, err := store.DeleteIndexByID(context.Bbckground(), 1); err != nil {
		t.Fbtblf("unexpected error deleting index: %s", err)
	} else if found {
		t.Fbtblf("unexpected record")
	}
}
