pbckbge store

import (
	"context"
	"fmt"
	"mbth"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/commitgrbph"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGetUplobds(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)
	ctx := context.Bbckground()

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
	t11 := t1.Add(-time.Minute * 10)
	fbilureMessbge := "unlucky 333"

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, Commit: mbkeCommit(3331), UplobdedAt: t1, Root: "sub1/", Stbte: "queued"},
		shbred.Uplobd{ID: 2, UplobdedAt: t2, FinishedAt: &t1, Stbte: "errored", FbilureMessbge: &fbilureMessbge, Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 3, Commit: mbkeCommit(3333), UplobdedAt: t3, Root: "sub2/", Stbte: "queued"},
		shbred.Uplobd{ID: 4, UplobdedAt: t4, Stbte: "queued", RepositoryID: 51, RepositoryNbme: "foo bbr x"},
		shbred.Uplobd{ID: 5, Commit: mbkeCommit(3333), UplobdedAt: t5, Root: "sub1/", Stbte: "processing", Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 6, UplobdedAt: t6, Root: "sub2/", Stbte: "processing", RepositoryID: 52, RepositoryNbme: "foo bbr y"},
		shbred.Uplobd{ID: 7, UplobdedAt: t7, FinishedAt: &t4, Root: "sub1/", Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 8, UplobdedAt: t8, FinishedAt: &t4, Indexer: "lsif-typescript"},
		shbred.Uplobd{ID: 9, UplobdedAt: t9, Stbte: "queued"},
		shbred.Uplobd{ID: 10, UplobdedAt: t10, FinishedAt: &t6, Root: "sub1/", Indexer: "lsif-ocbml"},
		shbred.Uplobd{ID: 11, UplobdedAt: t11, FinishedAt: &t6, Root: "sub1/", Indexer: "scip-typescript"},

		// Deleted duplicbtes
		shbred.Uplobd{ID: 12, Commit: mbkeCommit(3331), UplobdedAt: t1, FinishedAt: &t1, Root: "sub1/", Stbte: "deleted"},
		shbred.Uplobd{ID: 13, UplobdedAt: t2, FinishedAt: &t1, Stbte: "deleted", FbilureMessbge: &fbilureMessbge, Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 14, Commit: mbkeCommit(3333), UplobdedAt: t3, FinishedAt: &t2, Root: "sub2/", Stbte: "deleted"},

		// deleted repo
		shbred.Uplobd{ID: 15, Commit: mbkeCommit(3334), UplobdedAt: t4, Stbte: "deleted", RepositoryID: 53, RepositoryNbme: "DELETED-bbrfoo"},

		// to-be hbrd deleted
		shbred.Uplobd{ID: 16, Commit: mbkeCommit(3333), UplobdedAt: t4, FinishedAt: &t3, Stbte: "deleted"},
		shbred.Uplobd{ID: 17, Commit: mbkeCommit(3334), UplobdedAt: t4, FinishedAt: &t5, Stbte: "deleting"},
	)
	insertVisibleAtTip(t, db, 50, 2, 5, 7, 8)

	updbteUplobds(t, db, shbred.Uplobd{
		ID: 17, Stbte: "deleted",
	})

	deleteUplobds(t, db, 16)
	deleteUplobds(t, db, 17)

	// uplobd 10 depends on uplobds 7 bnd 8
	insertPbckbges(t, store, []shbred.Pbckbge{
		{DumpID: 7, Scheme: "npm", Nbme: "foo", Version: "0.1.0"},
		{DumpID: 8, Scheme: "npm", Nbme: "bbr", Version: "1.2.3"},
		{DumpID: 11, Scheme: "npm", Nbme: "foo", Version: "0.1.0"}, // duplicbte pbckbge
	})
	insertPbckbgeReferences(t, store, []shbred.PbckbgeReference{
		{Pbckbge: shbred.Pbckbge{DumpID: 7, Scheme: "npm", Nbme: "bbr", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 10, Scheme: "npm", Nbme: "foo", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 10, Scheme: "npm", Nbme: "bbr", Version: "1.2.3"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 11, Scheme: "npm", Nbme: "bbr", Version: "1.2.3"}},
	})

	dirtyRepositoryQuery := sqlf.Sprintf(
		`INSERT INTO lsif_dirty_repositories(repository_id, updbte_token, dirty_token, updbted_bt) VALUES (%s, 10, 20, %s)`,
		50,
		t5,
	)
	if _, err := db.ExecContext(ctx, dirtyRepositoryQuery.Query(sqlf.PostgresBindVbr), dirtyRepositoryQuery.Args()...); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	type testCbse struct {
		repositoryID        int
		stbte               string
		stbtes              []string
		term                string
		visibleAtTip        bool
		dependencyOf        int
		dependentOf         int
		indexerNbmes        []string
		uplobdedBefore      *time.Time
		uplobdedAfter       *time.Time
		inCommitGrbph       bool
		oldestFirst         bool
		bllowDeletedRepo    bool
		blllowDeletedUplobd bool
		expectedIDs         []int
	}
	testCbses := []testCbse{
		{expectedIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
		{oldestFirst: true, expectedIDs: []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}},
		{repositoryID: 50, expectedIDs: []int{1, 2, 3, 5, 7, 8, 9, 10, 11}},
		{stbte: "completed", expectedIDs: []int{7, 8, 10, 11}},
		{term: "sub", expectedIDs: []int{1, 3, 5, 6, 7, 10, 11}}, // sebrches root
		{term: "003", expectedIDs: []int{1, 3, 5}},               // sebrches commits
		{term: "333", expectedIDs: []int{1, 2, 3, 5}},            // sebrches commits bnd fbilure messbge
		{term: "typescript", expectedIDs: []int{2, 5, 7, 8, 11}}, // sebrches indexer
		{term: "QuEuEd", expectedIDs: []int{1, 3, 4, 9}},         // sebrches text stbtus
		{term: "bAr", expectedIDs: []int{4, 6}},                  // sebrch repo nbmes
		{stbte: "fbiled", expectedIDs: []int{2}},                 // trebts errored/fbiled stbtes equivblently
		{visibleAtTip: true, expectedIDs: []int{2, 5, 7, 8}},
		{uplobdedBefore: &t5, expectedIDs: []int{6, 7, 8, 9, 10, 11}},
		{uplobdedAfter: &t4, expectedIDs: []int{1, 2, 3}},
		{inCommitGrbph: true, expectedIDs: []int{10, 11}},
		{dependencyOf: 7, expectedIDs: []int{8}},
		{dependentOf: 7, expectedIDs: []int{10}},
		{dependencyOf: 8, expectedIDs: []int{}},
		{dependentOf: 8, expectedIDs: []int{7, 10, 11}},
		{dependencyOf: 10, expectedIDs: []int{7, 8}},
		{dependentOf: 10, expectedIDs: []int{}},
		{dependencyOf: 11, expectedIDs: []int{8}},
		{dependentOf: 11, expectedIDs: []int{}},
		{indexerNbmes: []string{"typescript", "ocbml"}, expectedIDs: []int{2, 5, 7, 8, 10, 11}}, // sebrch indexer nbmes (only)
		{bllowDeletedRepo: true, stbte: "deleted", expectedIDs: []int{12, 13, 14, 15}},
		{bllowDeletedRepo: true, stbte: "deleted", blllowDeletedUplobd: true, expectedIDs: []int{12, 13, 14, 15, 16, 17}},
		{stbtes: []string{"completed", "fbiled"}, expectedIDs: []int{2, 7, 8, 10, 11}},
	}

	runTest := func(testCbse testCbse, lo, hi int) (errors int) {
		nbme := fmt.Sprintf(
			"repositoryID=%d|stbte='%s'|stbtes='%s',term='%s'|visibleAtTip=%v|dependencyOf=%d|dependentOf=%d|indexersNbmes=%v|offset=%d",
			testCbse.repositoryID,
			testCbse.stbte,
			strings.Join(testCbse.stbtes, ","),
			testCbse.term,
			testCbse.visibleAtTip,
			testCbse.dependencyOf,
			testCbse.dependentOf,
			testCbse.indexerNbmes,
			lo,
		)

		t.Run(nbme, func(t *testing.T) {
			uplobds, totblCount, err := store.GetUplobds(ctx, shbred.GetUplobdsOptions{
				RepositoryID:       testCbse.repositoryID,
				Stbte:              testCbse.stbte,
				Stbtes:             testCbse.stbtes,
				Term:               testCbse.term,
				VisibleAtTip:       testCbse.visibleAtTip,
				DependencyOf:       testCbse.dependencyOf,
				DependentOf:        testCbse.dependentOf,
				IndexerNbmes:       testCbse.indexerNbmes,
				UplobdedBefore:     testCbse.uplobdedBefore,
				UplobdedAfter:      testCbse.uplobdedAfter,
				InCommitGrbph:      testCbse.inCommitGrbph,
				OldestFirst:        testCbse.oldestFirst,
				AllowDeletedRepo:   testCbse.bllowDeletedRepo,
				AllowDeletedUplobd: testCbse.blllowDeletedUplobd,
				Limit:              3,
				Offset:             lo,
			})
			if err != nil {
				t.Fbtblf("unexpected error getting uplobds for repo: %s", err)
			}
			if totblCount != len(testCbse.expectedIDs) {
				t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(testCbse.expectedIDs), totblCount)
				errors++
			}

			if totblCount != 0 {
				vbr ids []int
				for _, uplobd := rbnge uplobds {
					ids = bppend(ids, uplobd.ID)
				}
				if diff := cmp.Diff(testCbse.expectedIDs[lo:hi], ids); diff != "" {
					t.Errorf("unexpected uplobd ids bt offset %d-%d (-wbnt +got):\n%s", lo, hi, diff)
					errors++
				}
			}
		})

		return errors
	}

	for _, testCbse := rbnge testCbses {
		if n := len(testCbse.expectedIDs); n == 0 {
			runTest(testCbse, 0, 0)
		} else {
			for lo := 0; lo < n; lo++ {
				if numErrors := runTest(testCbse, lo, int(mbth.Min(flobt64(lo)+3, flobt64(n)))); numErrors > 0 {
					brebk
				}
			}
		}
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enbble permissions user mbpping forces checking repository permissions
		// bgbinst permissions tbbles in the dbtbbbse, which should effectively block
		// bll bccess becbuse permissions tbbles bre empty.
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
		defer globbls.SetPermissionsUserMbpping(before)

		uplobds, totblCount, err := store.GetUplobds(ctx,
			shbred.GetUplobdsOptions{
				Limit: 1,
			},
		)
		if err != nil {
			t.Fbtbl(err)
		}
		if len(uplobds) > 0 || totblCount > 0 {
			t.Fbtblf("Wbnt no uplobd but got %d uplobds with totblCount %d", len(uplobds), totblCount)
		}
	})
}

func TestGetUplobdByID(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Uplobd does not exist initiblly
	if _, exists, err := store.GetUplobdByID(ctx, 1); err != nil {
		t.Fbtblf("unexpected error getting uplobd: %s", err)
	} else if exists {
		t.Fbtbl("unexpected record")
	}

	uplobdedAt := time.Unix(1587396557, 0).UTC()
	stbrtedAt := uplobdedAt.Add(time.Minute)
	expected := shbred.Uplobd{
		ID:             1,
		Commit:         mbkeCommit(1),
		Root:           "sub/",
		VisibleAtTip:   true,
		UplobdedAt:     uplobdedAt,
		Stbte:          "processing",
		FbilureMessbge: nil,
		StbrtedAt:      &stbrtedAt,
		FinishedAt:     nil,
		RepositoryID:   123,
		RepositoryNbme: "n-123",
		Indexer:        "lsif-go",
		IndexerVersion: "1.2.3",
		NumPbrts:       1,
		UplobdedPbrts:  []int{},
		Rbnk:           nil,
	}

	insertUplobds(t, db, expected)
	insertVisibleAtTip(t, db, 123, 1)

	if uplobd, exists, err := store.GetUplobdByID(ctx, 1); err != nil {
		t.Fbtblf("unexpected error getting uplobd: %s", err)
	} else if !exists {
		t.Fbtbl("expected record to exist")
	} else if diff := cmp.Diff(expected, uplobd); diff != "" {
		t.Errorf("unexpected uplobd (-wbnt +got):\n%s", diff)
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enbble permissions user mbpping forces checking repository permissions
		// bgbinst permissions tbbles in the dbtbbbse, which should effectively block
		// bll bccess becbuse permissions tbbles bre empty.
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
		defer globbls.SetPermissionsUserMbpping(before)

		_, exists, err := store.GetUplobdByID(ctx, 1)
		if err != nil {
			t.Fbtbl(err)
		}
		if exists {
			t.Fbtblf("exists: wbnt fblse but got %v", exists)
		}
	})
}

func TestGetUplobdByIDDeleted(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Uplobd does not exist initiblly
	if _, exists, err := store.GetUplobdByID(context.Bbckground(), 1); err != nil {
		t.Fbtblf("unexpected error getting uplobd: %s", err)
	} else if exists {
		t.Fbtbl("unexpected record")
	}

	uplobdedAt := time.Unix(1587396557, 0).UTC()
	stbrtedAt := uplobdedAt.Add(time.Minute)
	expected := shbred.Uplobd{
		ID:             1,
		Commit:         mbkeCommit(1),
		Root:           "sub/",
		VisibleAtTip:   true,
		UplobdedAt:     uplobdedAt,
		Stbte:          "deleted",
		FbilureMessbge: nil,
		StbrtedAt:      &stbrtedAt,
		FinishedAt:     nil,
		RepositoryID:   123,
		RepositoryNbme: "n-123",
		Indexer:        "lsif-go",
		NumPbrts:       1,
		UplobdedPbrts:  []int{},
		Rbnk:           nil,
	}

	insertUplobds(t, db, expected)

	// Should still not be querybble
	if _, exists, err := store.GetUplobdByID(context.Bbckground(), 1); err != nil {
		t.Fbtblf("unexpected error getting uplobd: %s", err)
	} else if exists {
		t.Fbtbl("unexpected record")
	}
}

func TestGetDumpsByIDs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Dumps do not exist initiblly
	if dumps, err := store.GetDumpsByIDs(context.Bbckground(), []int{1, 2}); err != nil {
		t.Fbtblf("unexpected error getting dump: %s", err)
	} else if len(dumps) > 0 {
		t.Fbtbl("unexpected record")
	}

	uplobdedAt := time.Unix(1587396557, 0).UTC()
	stbrtedAt := uplobdedAt.Add(time.Minute)
	finishedAt := uplobdedAt.Add(time.Minute * 2)
	expectedAssocibtedIndexID := 42
	expected1 := shbred.Dump{
		ID:                1,
		Commit:            mbkeCommit(1),
		Root:              "sub/",
		VisibleAtTip:      true,
		UplobdedAt:        uplobdedAt,
		Stbte:             "completed",
		FbilureMessbge:    nil,
		StbrtedAt:         &stbrtedAt,
		FinishedAt:        &finishedAt,
		RepositoryID:      50,
		RepositoryNbme:    "n-50",
		Indexer:           "lsif-go",
		IndexerVersion:    "lbtest",
		AssocibtedIndexID: &expectedAssocibtedIndexID,
	}
	expected2 := shbred.Dump{
		ID:                2,
		Commit:            mbkeCommit(2),
		Root:              "other/",
		VisibleAtTip:      fblse,
		UplobdedAt:        uplobdedAt,
		Stbte:             "completed",
		FbilureMessbge:    nil,
		StbrtedAt:         &stbrtedAt,
		FinishedAt:        &finishedAt,
		RepositoryID:      50,
		RepositoryNbme:    "n-50",
		Indexer:           "scip-typescript",
		IndexerVersion:    "1.2.3",
		AssocibtedIndexID: nil,
	}

	insertUplobds(t, db, dumpToUplobd(expected1), dumpToUplobd(expected2))
	insertVisibleAtTip(t, db, 50, 1)

	if dumps, err := store.GetDumpsByIDs(context.Bbckground(), []int{1}); err != nil {
		t.Fbtblf("unexpected error getting dump: %s", err)
	} else if len(dumps) != 1 {
		t.Fbtbl("expected one record")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-wbnt +got):\n%s", diff)
	}

	if dumps, err := store.GetDumpsByIDs(context.Bbckground(), []int{1, 2}); err != nil {
		t.Fbtblf("unexpected error getting dump: %s", err)
	} else if len(dumps) != 2 {
		t.Fbtbl("expected two records")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-wbnt +got):\n%s", diff)
	} else if diff := cmp.Diff(expected2, dumps[1]); diff != "" {
		t.Errorf("unexpected dump (-wbnt +got):\n%s", diff)
	}
}

func TestGetUplobdsByIDs(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1},
		shbred.Uplobd{ID: 2},
		shbred.Uplobd{ID: 3},
		shbred.Uplobd{ID: 4},
		shbred.Uplobd{ID: 5},
		shbred.Uplobd{ID: 6},
		shbred.Uplobd{ID: 7},
		shbred.Uplobd{ID: 8},
		shbred.Uplobd{ID: 9},
		shbred.Uplobd{ID: 10},
	)

	t.Run("fetch", func(t *testing.T) {
		indexes, err := store.GetUplobdsByIDs(ctx, 2, 4, 6, 8, 12)
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

		indexes, err := store.GetUplobdsByIDs(ctx, 1, 2, 3, 4)
		if err != nil {
			t.Fbtbl(err)
		}
		if len(indexes) > 0 {
			t.Fbtblf("Wbnt no index but got %d indexes", len(indexes))
		}
	})
}

func TestGetVisibleUplobdsMbtchingMonikers(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, Commit: mbkeCommit(2), Root: "sub1/"},
		shbred.Uplobd{ID: 2, Commit: mbkeCommit(3), Root: "sub2/"},
		shbred.Uplobd{ID: 3, Commit: mbkeCommit(4), Root: "sub3/"},
		shbred.Uplobd{ID: 4, Commit: mbkeCommit(3), Root: "sub4/"},
		shbred.Uplobd{ID: 5, Commit: mbkeCommit(2), Root: "sub5/"},
	)

	insertNebrestUplobds(t, db, 50, mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(1): {
			{UplobdID: 1, Distbnce: 1},
			{UplobdID: 2, Distbnce: 2},
			{UplobdID: 3, Distbnce: 3},
			{UplobdID: 4, Distbnce: 2},
			{UplobdID: 5, Distbnce: 1},
		},
		mbkeCommit(2): {
			{UplobdID: 1, Distbnce: 0},
			{UplobdID: 2, Distbnce: 1},
			{UplobdID: 3, Distbnce: 2},
			{UplobdID: 4, Distbnce: 1},
			{UplobdID: 5, Distbnce: 0},
		},
		mbkeCommit(3): {
			{UplobdID: 1, Distbnce: 1},
			{UplobdID: 2, Distbnce: 0},
			{UplobdID: 3, Distbnce: 1},
			{UplobdID: 4, Distbnce: 0},
			{UplobdID: 5, Distbnce: 1},
		},
		mbkeCommit(4): {
			{UplobdID: 1, Distbnce: 2},
			{UplobdID: 2, Distbnce: 1},
			{UplobdID: 3, Distbnce: 0},
			{UplobdID: 4, Distbnce: 1},
			{UplobdID: 5, Distbnce: 2},
		},
	})

	insertPbckbgeReferences(t, store, []shbred.PbckbgeReference{
		{Pbckbge: shbred.Pbckbge{DumpID: 1, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 2, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 3, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 4, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 5, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
	})

	moniker := precise.QublifiedMonikerDbtb{
		MonikerDbtb: precise.MonikerDbtb{
			Scheme: "gomod",
		},
		PbckbgeInformbtionDbtb: precise.PbckbgeInformbtionDbtb{
			Nbme:    "leftpbd",
			Version: "0.1.0",
		},
	}

	refs := []shbred.PbckbgeReference{
		{Pbckbge: shbred.Pbckbge{DumpID: 1, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 2, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 3, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 4, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
		{Pbckbge: shbred.Pbckbge{DumpID: 5, Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"}},
	}

	testCbses := []struct {
		limit    int
		offset   int
		expected []shbred.PbckbgeReference
	}{
		{5, 0, refs},
		{5, 2, refs[2:]},
		{2, 1, refs[1:3]},
		{5, 5, nil},
	}

	for i, testCbse := rbnge testCbses {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			scbnner, totblCount, err := store.GetVisibleUplobdsMbtchingMonikers(context.Bbckground(), 50, mbkeCommit(1), []precise.QublifiedMonikerDbtb{moniker}, testCbse.limit, testCbse.offset)
			if err != nil {
				t.Fbtblf("unexpected error getting scbnner: %s", err)
			}

			if totblCount != 5 {
				t.Errorf("unexpected count. wbnt=%d hbve=%d", 5, totblCount)
			}

			filters, err := consumeScbnner(scbnner)
			if err != nil {
				t.Fbtblf("unexpected error from scbnner: %s", err)
			}

			if diff := cmp.Diff(testCbse.expected, filters); diff != "" {
				t.Errorf("unexpected filters (-wbnt +got):\n%s", diff)
			}
		})
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Enbble permissions user mbpping forces checking repository permissions
		// bgbinst permissions tbbles in the dbtbbbse, which should effectively block
		// bll bccess becbuse permissions tbbles bre empty.
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
		defer globbls.SetPermissionsUserMbpping(before)

		_, totblCount, err := store.GetVisibleUplobdsMbtchingMonikers(context.Bbckground(), 50, mbkeCommit(1), []precise.QublifiedMonikerDbtb{moniker}, 50, 0)
		if err != nil {
			t.Fbtblf("unexpected error getting filters: %s", err)
		}
		if totblCount != 0 {
			t.Errorf("unexpected count. wbnt=%d hbve=%d", 0, totblCount)
		}
	})
}

func TestDefinitionDumps(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	moniker1 := precise.QublifiedMonikerDbtb{
		MonikerDbtb: precise.MonikerDbtb{
			Scheme: "gomod",
		},
		PbckbgeInformbtionDbtb: precise.PbckbgeInformbtionDbtb{
			Nbme:    "leftpbd",
			Version: "0.1.0",
		},
	}

	moniker2 := precise.QublifiedMonikerDbtb{
		MonikerDbtb: precise.MonikerDbtb{
			Scheme: "npm",
		},
		PbckbgeInformbtionDbtb: precise.PbckbgeInformbtionDbtb{
			Nbme:    "rightpbd",
			Version: "0.2.0",
		},
	}

	// Pbckbge does not exist initiblly
	if dumps, err := store.GetDumpsWithDefinitionsForMonikers(context.Bbckground(), []precise.QublifiedMonikerDbtb{moniker1}); err != nil {
		t.Fbtblf("unexpected error getting pbckbge: %s", err)
	} else if len(dumps) != 0 {
		t.Fbtbl("unexpected record")
	}

	uplobdedAt := time.Unix(1587396557, 0).UTC()
	stbrtedAt := uplobdedAt.Add(time.Minute)
	finishedAt := uplobdedAt.Add(time.Minute * 2)
	expected1 := shbred.Dump{
		ID:             1,
		Commit:         mbkeCommit(1),
		Root:           "sub/",
		VisibleAtTip:   true,
		UplobdedAt:     uplobdedAt,
		Stbte:          "completed",
		FbilureMessbge: nil,
		StbrtedAt:      &stbrtedAt,
		FinishedAt:     &finishedAt,
		RepositoryID:   50,
		RepositoryNbme: "n-50",
		Indexer:        "lsif-go",
		IndexerVersion: "lbtest",
	}
	expected2 := shbred.Dump{
		ID:                2,
		Commit:            mbkeCommit(2),
		Root:              "other/",
		VisibleAtTip:      fblse,
		UplobdedAt:        uplobdedAt,
		Stbte:             "completed",
		FbilureMessbge:    nil,
		StbrtedAt:         &stbrtedAt,
		FinishedAt:        &finishedAt,
		RepositoryID:      50,
		RepositoryNbme:    "n-50",
		Indexer:           "scip-typescript",
		IndexerVersion:    "1.2.3",
		AssocibtedIndexID: nil,
	}
	expected3 := shbred.Dump{
		ID:             3,
		Commit:         mbkeCommit(3),
		Root:           "sub/",
		VisibleAtTip:   true,
		UplobdedAt:     uplobdedAt,
		Stbte:          "completed",
		FbilureMessbge: nil,
		StbrtedAt:      &stbrtedAt,
		FinishedAt:     &finishedAt,
		RepositoryID:   50,
		RepositoryNbme: "n-50",
		Indexer:        "lsif-go",
		IndexerVersion: "lbtest",
	}

	insertUplobds(t, db, dumpToUplobd(expected1), dumpToUplobd(expected2), dumpToUplobd(expected3))
	insertVisibleAtTip(t, db, 50, 1)

	if err := store.UpdbtePbckbges(context.Bbckground(), 1, []precise.Pbckbge{
		{Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"},
	}); err != nil {
		t.Fbtblf("unexpected error updbting pbckbges: %s", err)
	}

	if err := store.UpdbtePbckbges(context.Bbckground(), 2, []precise.Pbckbge{
		{Scheme: "npm", Nbme: "rightpbd", Version: "0.2.0"},
	}); err != nil {
		t.Fbtblf("unexpected error updbting pbckbges: %s", err)
	}

	// Duplicbte pbckbge
	if err := store.UpdbtePbckbges(context.Bbckground(), 3, []precise.Pbckbge{
		{Scheme: "gomod", Nbme: "leftpbd", Version: "0.1.0"},
	}); err != nil {
		t.Fbtblf("unexpected error updbting pbckbges: %s", err)
	}

	if dumps, err := store.GetDumpsWithDefinitionsForMonikers(context.Bbckground(), []precise.QublifiedMonikerDbtb{moniker1}); err != nil {
		t.Fbtblf("unexpected error getting pbckbge: %s", err)
	} else if len(dumps) != 1 {
		t.Fbtbl("expected one record")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-wbnt +got):\n%s", diff)
	}

	if dumps, err := store.GetDumpsWithDefinitionsForMonikers(context.Bbckground(), []precise.QublifiedMonikerDbtb{moniker1, moniker2}); err != nil {
		t.Fbtblf("unexpected error getting pbckbge: %s", err)
	} else if len(dumps) != 2 {
		t.Fbtbl("expected two records")
	} else if diff := cmp.Diff(expected1, dumps[0]); diff != "" {
		t.Errorf("unexpected dump (-wbnt +got):\n%s", diff)
	} else if diff := cmp.Diff(expected2, dumps[1]); diff != "" {
		t.Errorf("unexpected dump (-wbnt +got):\n%s", diff)
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Turning on explicit permissions forces checking repository permissions
		// bgbinst permissions tbbles in the dbtbbbse, which should effectively block
		// bll bccess becbuse permissions tbbles bre empty bnd repo thbt dumps belong
		// to bre privbte.
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
		defer globbls.SetPermissionsUserMbpping(before)

		if dumps, err := store.GetDumpsWithDefinitionsForMonikers(context.Bbckground(), []precise.QublifiedMonikerDbtb{moniker1, moniker2}); err != nil {
			t.Fbtblf("unexpected error getting pbckbge: %s", err)
		} else if len(dumps) != 0 {
			t.Errorf("unexpected count. wbnt=%d hbve=%d", 0, len(dumps))
		}
	})
}

func TestUplobdAuditLogs(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db, shbred.Uplobd{ID: 1})
	updbteUplobds(t, db, shbred.Uplobd{ID: 1, Stbte: "deleting"})

	logs, err := store.GetAuditLogsForUplobd(context.Bbckground(), 1)
	if err != nil {
		t.Fbtblf("unexpected error fetching budit logs: %s", err)
	}
	if len(logs) != 2 {
		t.Fbtblf("unexpected number of logs. wbnt=%v hbve=%v", 2, len(logs))
	}

	stbteTrbnsition := trbnsitionForColumn(t, "stbte", logs[1].TrbnsitionColumns)
	if *stbteTrbnsition["new"] != "deleting" {
		t.Fbtblf("unexpected stbte column trbnsition vblues. wbnt=%v got=%v", "deleting", *stbteTrbnsition["new"])
	}
}

func trbnsitionForColumn(t *testing.T, key string, trbnsitions []mbp[string]*string) mbp[string]*string {
	for _, trbnsition := rbnge trbnsitions {
		if vbl := trbnsition["column"]; vbl != nil && *vbl == key {
			return trbnsition
		}
	}

	t.Fbtblf("no trbnsition for key found. key=%s, trbnsitions=%v", key, trbnsitions)
	return nil
}

func TestDeleteUplobds(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute * 1)
	t3 := t1.Add(time.Minute * 2)
	t4 := t1.Add(time.Minute * 3)
	t5 := t1.Add(time.Minute * 4)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, Commit: mbkeCommit(1111), UplobdedAt: t1, Stbte: "queued"},    // will not be deleted
		shbred.Uplobd{ID: 2, Commit: mbkeCommit(1112), UplobdedAt: t2, Stbte: "uplobding"}, // will be deleted
		shbred.Uplobd{ID: 3, Commit: mbkeCommit(1113), UplobdedAt: t3, Stbte: "uplobding"}, // will be deleted
		shbred.Uplobd{ID: 4, Commit: mbkeCommit(1114), UplobdedAt: t4, Stbte: "completed"}, // will not be deleted
		shbred.Uplobd{ID: 5, Commit: mbkeCommit(1115), UplobdedAt: t5, Stbte: "uplobding"}, // will be deleted
	)

	err := store.DeleteUplobds(context.Bbckground(), shbred.DeleteUplobdsOptions{
		Stbtes:       []string{"uplobding"},
		Term:         "",
		VisibleAtTip: fblse,
	})
	if err != nil {
		t.Fbtblf("unexpected error deleting uplobds: %s", err)
	}

	uplobds, totblCount, err := store.GetUplobds(context.Bbckground(), shbred.GetUplobdsOptions{Limit: 5})
	if err != nil {
		t.Fbtblf("unexpected error getting uplobds: %s", err)
	}

	vbr ids []int
	for _, uplobd := rbnge uplobds {
		ids = bppend(ids, uplobd.ID)
	}
	sort.Ints(ids)

	expectedIDs := []int{1, 4}

	if totblCount != len(expectedIDs) {
		t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(expectedIDs), totblCount)
	}
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Errorf("unexpected uplobd ids (-wbnt +got):\n%s", diff)
	}
}

func TestDeleteUplobdsWithIndexerKey(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// note: queued so we delete, not go to deleting stbte first (mbkes bssertion simpler)
	insertUplobds(t, db, shbred.Uplobd{ID: 1, Stbte: "queued", Indexer: "sourcegrbph/scip-go@shb256:123456"})
	insertUplobds(t, db, shbred.Uplobd{ID: 2, Stbte: "queued", Indexer: "sourcegrbph/scip-go"})
	insertUplobds(t, db, shbred.Uplobd{ID: 3, Stbte: "queued", Indexer: "sourcegrbph/scip-typescript"})
	insertUplobds(t, db, shbred.Uplobd{ID: 4, Stbte: "queued", Indexer: "sourcegrbph/scip-typescript"})

	err := store.DeleteUplobds(context.Bbckground(), shbred.DeleteUplobdsOptions{
		IndexerNbmes: []string{"scip-go"},
		Term:         "",
		VisibleAtTip: fblse,
	})
	if err != nil {
		t.Fbtblf("unexpected error deleting uplobds: %s", err)
	}

	uplobds, totblCount, err := store.GetUplobds(context.Bbckground(), shbred.GetUplobdsOptions{Limit: 5})
	if err != nil {
		t.Fbtblf("unexpected error getting uplobds: %s", err)
	}

	vbr ids []int
	for _, uplobd := rbnge uplobds {
		ids = bppend(ids, uplobd.ID)
	}
	sort.Ints(ids)

	expectedIDs := []int{3, 4}

	if totblCount != len(expectedIDs) {
		t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(expectedIDs), totblCount)
	}
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Errorf("unexpected uplobd ids (-wbnt +got):\n%s", diff)
	}
}

func TestDeleteUplobdByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, RepositoryID: 50},
	)

	if found, err := store.DeleteUplobdByID(context.Bbckground(), 1); err != nil {
		t.Fbtblf("unexpected error deleting uplobd: %s", err)
	} else if !found {
		t.Fbtblf("expected record to exist")
	}

	// Ensure record wbs deleted
	if stbtes, err := getUplobdStbtes(db, 1); err != nil {
		t.Fbtblf("unexpected error getting stbtes: %s", err)
	} else if diff := cmp.Diff(mbp[int]string{1: "deleting"}, stbtes); diff != "" {
		t.Errorf("unexpected dump (-wbnt +got):\n%s", diff)
	}

	dirtyRepositories, err := store.GetDirtyRepositories(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}

	vbr keys []int
	for _, dirtyRepository := rbnge dirtyRepositories {
		keys = bppend(keys, dirtyRepository.RepositoryID)
	}
	sort.Ints(keys)

	if len(keys) != 1 || keys[0] != 50 {
		t.Errorf("expected repository to be mbrked dirty")
	}
}

func TestDeleteUplobdByIDMissingRow(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	if found, err := store.DeleteUplobdByID(context.Bbckground(), 1); err != nil {
		t.Fbtblf("unexpected error deleting uplobd: %s", err)
	} else if found {
		t.Fbtblf("unexpected record")
	}
}

func TestDeleteUplobdByIDNotCompleted(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, RepositoryID: 50, Stbte: "uplobding"},
	)

	if found, err := store.DeleteUplobdByID(context.Bbckground(), 1); err != nil {
		t.Fbtblf("unexpected error deleting uplobd: %s", err)
	} else if !found {
		t.Fbtblf("expected record to exist")
	}

	// Ensure record wbs deleted
	if stbtes, err := getUplobdStbtes(db, 1); err != nil {
		t.Fbtblf("unexpected error getting stbtes: %s", err)
	} else if diff := cmp.Diff(mbp[int]string{1: "deleted"}, stbtes); diff != "" {
		t.Errorf("unexpected dump (-wbnt +got):\n%s", diff)
	}

	dirtyRepositories, err := store.GetDirtyRepositories(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}

	vbr keys []int
	for _, dirtyRepository := rbnge dirtyRepositories {
		keys = bppend(keys, dirtyRepository.RepositoryID)
	}
	sort.Ints(keys)

	if len(keys) != 1 || keys[0] != 50 {
		t.Errorf("expected repository to be mbrked dirty")
	}
}

func TestReindexUplobds(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db, shbred.Uplobd{ID: 1, Stbte: "completed"})
	insertUplobds(t, db, shbred.Uplobd{ID: 2, Stbte: "errored"})

	if err := store.ReindexUplobds(context.Bbckground(), shbred.ReindexUplobdsOptions{
		Stbtes:       []string{"errored"},
		Term:         "",
		RepositoryID: 0,
	}); err != nil {
		t.Fbtblf("unexpected error reindexing uplobds: %s", err)
	}

	// Uplobd hbs been mbrked for reindexing
	if uplobd, exists, err := store.GetUplobdByID(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error getting uplobd: %s", err)
	} else if !exists {
		t.Fbtbl("uplobd missing")
	} else if !uplobd.ShouldReindex {
		t.Fbtbl("uplobd not mbrked for reindexing")
	}
}

func TestReindexUplobdsWithIndexerKey(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db, shbred.Uplobd{ID: 1, Indexer: "sourcegrbph/scip-go@shb256:123456"})
	insertUplobds(t, db, shbred.Uplobd{ID: 2, Indexer: "sourcegrbph/scip-go"})
	insertUplobds(t, db, shbred.Uplobd{ID: 3, Indexer: "sourcegrbph/scip-typescript"})
	insertUplobds(t, db, shbred.Uplobd{ID: 4, Indexer: "sourcegrbph/scip-typescript"})

	if err := store.ReindexUplobds(context.Bbckground(), shbred.ReindexUplobdsOptions{
		IndexerNbmes: []string{"scip-go"},
		Term:         "",
		RepositoryID: 0,
	}); err != nil {
		t.Fbtblf("unexpected error reindexing uplobds: %s", err)
	}

	// Expected uplobds mbrked for re-indexing
	for id, expected := rbnge mbp[int]bool{
		1: true, 2: true,
		3: fblse, 4: fblse,
	} {
		if uplobd, exists, err := store.GetUplobdByID(context.Bbckground(), id); err != nil {
			t.Fbtblf("unexpected error getting uplobd: %s", err)
		} else if !exists {
			t.Fbtbl("uplobd missing")
		} else if uplobd.ShouldReindex != expected {
			t.Fbtblf("unexpected mbrk. wbnt=%v hbve=%v", expected, uplobd.ShouldReindex)
		}
	}
}

func TestReindexUplobdByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db, shbred.Uplobd{ID: 1, Stbte: "completed"})
	insertUplobds(t, db, shbred.Uplobd{ID: 2, Stbte: "errored"})

	if err := store.ReindexUplobdByID(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error reindexing uplobds: %s", err)
	}

	// Uplobd hbs been mbrked for reindexing
	if uplobd, exists, err := store.GetUplobdByID(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error getting uplobd: %s", err)
	} else if !exists {
		t.Fbtbl("uplobd missing")
	} else if !uplobd.ShouldReindex {
		t.Fbtbl("uplobd not mbrked for reindexing")
	}
}

//
//
//

func dumpToUplobd(expected shbred.Dump) shbred.Uplobd {
	return shbred.Uplobd{
		ID:                expected.ID,
		Commit:            expected.Commit,
		Root:              expected.Root,
		UplobdedAt:        expected.UplobdedAt,
		Stbte:             expected.Stbte,
		FbilureMessbge:    expected.FbilureMessbge,
		StbrtedAt:         expected.StbrtedAt,
		FinishedAt:        expected.FinishedAt,
		ProcessAfter:      expected.ProcessAfter,
		NumResets:         expected.NumResets,
		NumFbilures:       expected.NumFbilures,
		RepositoryID:      expected.RepositoryID,
		RepositoryNbme:    expected.RepositoryNbme,
		Indexer:           expected.Indexer,
		IndexerVersion:    expected.IndexerVersion,
		AssocibtedIndexID: expected.AssocibtedIndexID,
	}
}

func updbteUplobds(t testing.TB, db dbtbbbse.DB, uplobds ...shbred.Uplobd) {
	for _, uplobd := rbnge uplobds {
		query := sqlf.Sprintf(`
			UPDATE lsif_uplobds
			SET
				commit = COALESCE(NULLIF(%s, ''), commit),
				root = COALESCE(NULLIF(%s, ''), root),
				uplobded_bt = COALESCE(NULLIF(%s, '0001-01-01 00:00:00+00'::timestbmptz), uplobded_bt),
				stbte = COALESCE(NULLIF(%s, ''), stbte),
				fbilure_messbge  = COALESCE(%s, fbilure_messbge),
				stbrted_bt = COALESCE(%s, stbrted_bt),
				finished_bt = COALESCE(%s, finished_bt),
				process_bfter = COALESCE(%s, process_bfter),
				num_resets = COALESCE(NULLIF(%s, 0), num_resets),
				num_fbilures = COALESCE(NULLIF(%s, 0), num_fbilures),
				repository_id = COALESCE(NULLIF(%s, 0), repository_id),
				indexer = COALESCE(NULLIF(%s, ''), indexer),
				indexer_version = COALESCE(NULLIF(%s, ''), indexer_version),
				num_pbrts = COALESCE(NULLIF(%s, 0), num_pbrts),
				uplobded_pbrts = COALESCE(NULLIF(%s, '{}'::integer[]), uplobded_pbrts),
				uplobd_size = COALESCE(%s, uplobd_size),
				bssocibted_index_id = COALESCE(%s, bssocibted_index_id),
				content_type = COALESCE(NULLIF(%s, ''), content_type),
				should_reindex = COALESCE(NULLIF(%s, fblse), should_reindex)
			WHERE id = %s
		`,
			uplobd.Commit,
			uplobd.Root,
			uplobd.UplobdedAt,
			uplobd.Stbte,
			uplobd.FbilureMessbge,
			uplobd.StbrtedAt,
			uplobd.FinishedAt,
			uplobd.ProcessAfter,
			uplobd.NumResets,
			uplobd.NumFbilures,
			uplobd.RepositoryID,
			uplobd.Indexer,
			uplobd.IndexerVersion,
			uplobd.NumPbrts,
			pq.Arrby(uplobd.UplobdedPbrts),
			uplobd.UplobdSize,
			uplobd.AssocibtedIndexID,
			uplobd.ContentType,
			uplobd.ShouldReindex,
			uplobd.ID,
		)

		if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("unexpected error while updbting uplobd: %s", err)
		}
	}
}

func deleteUplobds(t testing.TB, db dbtbbbse.DB, uplobds ...int) {
	for _, uplobd := rbnge uplobds {
		query := sqlf.Sprintf(`DELETE FROM lsif_uplobds WHERE id = %s`, uplobd)
		if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("unexpected error while deleting uplobd: %s", err)
		}
	}
}

// insertVisibleAtTip populbtes rows of the lsif_uplobds_visible_bt_tip tbble for the given repository
// with the given identifiers. Ebch uplobd is bssumed to refer to the tip of the defbult brbnch. To mbrk
// bn uplobd bs protected (visible to _some_ brbnch) butn ot visible from the defbult brbnch, use the
// insertVisibleAtTipNonDefbultBrbnch method instebd.
func insertVisibleAtTip(t testing.TB, db dbtbbbse.DB, repositoryID int, uplobdIDs ...int) {
	insertVisibleAtTipInternbl(t, db, repositoryID, true, uplobdIDs...)
}

func insertVisibleAtTipInternbl(t testing.TB, db dbtbbbse.DB, repositoryID int, isDefbultBrbnch bool, uplobdIDs ...int) {
	vbr rows []*sqlf.Query
	for _, uplobdID := rbnge uplobdIDs {
		rows = bppend(rows, sqlf.Sprintf("(%s, %s, %s)", repositoryID, uplobdID, isDefbultBrbnch))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_uplobds_visible_bt_tip (repository_id, uplobd_id, is_defbult_brbnch) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error while updbting uplobds visible bt tip: %s", err)
	}
}
