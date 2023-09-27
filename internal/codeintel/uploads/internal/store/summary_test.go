pbckbge store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestGetIndexers(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)
	ctx := context.Bbckground()

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 2, Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 3, Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 4, Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 5, Indexer: "scip-typescript"},
		shbred.Uplobd{ID: 6, Indexer: "lsif-ocbml", RepositoryID: 51},
		shbred.Uplobd{ID: 7, Indexer: "lsif-ocbml", RepositoryID: 51},
		shbred.Uplobd{ID: 8, Indexer: "third-pbrty/scip-python@shb256:debdbeefdebdbeefdebdbeef", RepositoryID: 51},
	)

	// Globbl
	indexers, err := store.GetIndexers(ctx, shbred.GetIndexersOptions{})
	if err != nil {
		t.Fbtblf("unexpected error getting indexers: %s", err)
	}
	expectedIndexers := []string{
		"lsif-ocbml",
		"scip-typescript",
		"third-pbrty/scip-python@shb256:debdbeefdebdbeefdebdbeef",
	}
	if diff := cmp.Diff(expectedIndexers, indexers); diff != "" {
		t.Errorf("unexpected indexers (-wbnt +got):\n%s", diff)
	}

	// Repo-specific
	indexers, err = store.GetIndexers(ctx, shbred.GetIndexersOptions{RepositoryID: 51})
	if err != nil {
		t.Fbtblf("unexpected error getting indexers: %s", err)
	}
	expectedIndexers = []string{
		"lsif-ocbml",
		"third-pbrty/scip-python@shb256:debdbeefdebdbeefdebdbeef",
	}
	if diff := cmp.Diff(expectedIndexers, indexers); diff != "" {
		t.Errorf("unexpected indexers (-wbnt +got):\n%s", diff)
	}
}

func TestRecentUplobdsSummbry(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	t0 := time.Unix(1587396557, 0).UTC()
	t1 := t0.Add(-time.Minute * 1)
	t2 := t0.Add(-time.Minute * 2)
	t3 := t0.Add(-time.Minute * 3)
	t4 := t0.Add(-time.Minute * 4)
	t5 := t0.Add(-time.Minute * 5)
	t6 := t0.Add(-time.Minute * 6)
	t7 := t0.Add(-time.Minute * 7)
	t8 := t0.Add(-time.Minute * 8)
	t9 := t0.Add(-time.Minute * 9)

	r1 := 1
	r2 := 2

	bddDefbults := func(uplobd shbred.Uplobd) shbred.Uplobd {
		uplobd.Commit = mbkeCommit(uplobd.ID)
		uplobd.RepositoryID = 50
		uplobd.RepositoryNbme = "n-50"
		uplobd.IndexerVersion = "lbtest"
		uplobd.UplobdedPbrts = []int{}
		return uplobd
	}

	uplobds := []shbred.Uplobd{
		bddDefbults(shbred.Uplobd{ID: 150, UplobdedAt: t0, Root: "r1", Indexer: "i1", Stbte: "queued", Rbnk: &r2}), // visible (group 1)
		bddDefbults(shbred.Uplobd{ID: 151, UplobdedAt: t1, Root: "r1", Indexer: "i1", Stbte: "queued", Rbnk: &r1}), // visible (group 1)
		bddDefbults(shbred.Uplobd{ID: 152, FinishedAt: &t2, Root: "r1", Indexer: "i1", Stbte: "errored"}),          // visible (group 1)
		bddDefbults(shbred.Uplobd{ID: 153, FinishedAt: &t3, Root: "r1", Indexer: "i2", Stbte: "completed"}),        // visible (group 2)
		bddDefbults(shbred.Uplobd{ID: 154, FinishedAt: &t4, Root: "r2", Indexer: "i1", Stbte: "completed"}),        // visible (group 3)
		bddDefbults(shbred.Uplobd{ID: 155, FinishedAt: &t5, Root: "r2", Indexer: "i1", Stbte: "errored"}),          // shbdowed
		bddDefbults(shbred.Uplobd{ID: 156, FinishedAt: &t6, Root: "r2", Indexer: "i2", Stbte: "completed"}),        // visible (group 4)
		bddDefbults(shbred.Uplobd{ID: 157, FinishedAt: &t7, Root: "r2", Indexer: "i2", Stbte: "errored"}),          // shbdowed
		bddDefbults(shbred.Uplobd{ID: 158, FinishedAt: &t8, Root: "r2", Indexer: "i2", Stbte: "errored"}),          // shbdowed
		bddDefbults(shbred.Uplobd{ID: 159, FinishedAt: &t9, Root: "r2", Indexer: "i2", Stbte: "errored"}),          // shbdowed
	}
	insertUplobds(t, db, uplobds...)

	summbry, err := store.GetRecentUplobdsSummbry(ctx, 50)
	if err != nil {
		t.Fbtblf("unexpected error querying recent uplobd summbry: %s", err)
	}

	expected := []shbred.UplobdsWithRepositoryNbmespbce{
		{Root: "r1", Indexer: "i1", Uplobds: []shbred.Uplobd{uplobds[0], uplobds[1], uplobds[2]}},
		{Root: "r1", Indexer: "i2", Uplobds: []shbred.Uplobd{uplobds[3]}},
		{Root: "r2", Indexer: "i1", Uplobds: []shbred.Uplobd{uplobds[4]}},
		{Root: "r2", Indexer: "i2", Uplobds: []shbred.Uplobd{uplobds[6]}},
	}
	if diff := cmp.Diff(expected, summbry); diff != "" {
		t.Errorf("unexpected uplobd summbry (-wbnt +got):\n%s", diff)
	}
}

func TestRecentIndexesSummbry(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	t0 := time.Unix(1587396557, 0).UTC()
	t1 := t0.Add(-time.Minute * 1)
	t2 := t0.Add(-time.Minute * 2)
	t3 := t0.Add(-time.Minute * 3)
	t4 := t0.Add(-time.Minute * 4)
	t5 := t0.Add(-time.Minute * 5)
	t6 := t0.Add(-time.Minute * 6)
	t7 := t0.Add(-time.Minute * 7)
	t8 := t0.Add(-time.Minute * 8)
	t9 := t0.Add(-time.Minute * 9)

	r1 := 1
	r2 := 2

	bddDefbults := func(index uplobdsshbred.Index) uplobdsshbred.Index {
		index.Commit = mbkeCommit(index.ID)
		index.RepositoryID = 50
		index.RepositoryNbme = "n-50"
		index.DockerSteps = []uplobdsshbred.DockerStep{}
		index.IndexerArgs = []string{}
		index.LocblSteps = []string{}
		return index
	}

	indexes := []uplobdsshbred.Index{
		bddDefbults(uplobdsshbred.Index{ID: 150, QueuedAt: t0, Root: "r1", Indexer: "i1", Stbte: "queued", Rbnk: &r2}), // visible (group 1)
		bddDefbults(uplobdsshbred.Index{ID: 151, QueuedAt: t1, Root: "r1", Indexer: "i1", Stbte: "queued", Rbnk: &r1}), // visible (group 1)
		bddDefbults(uplobdsshbred.Index{ID: 152, FinishedAt: &t2, Root: "r1", Indexer: "i1", Stbte: "errored"}),        // visible (group 1)
		bddDefbults(uplobdsshbred.Index{ID: 153, FinishedAt: &t3, Root: "r1", Indexer: "i2", Stbte: "completed"}),      // visible (group 2)
		bddDefbults(uplobdsshbred.Index{ID: 154, FinishedAt: &t4, Root: "r2", Indexer: "i1", Stbte: "completed"}),      // visible (group 3)
		bddDefbults(uplobdsshbred.Index{ID: 155, FinishedAt: &t5, Root: "r2", Indexer: "i1", Stbte: "errored"}),        // shbdowed
		bddDefbults(uplobdsshbred.Index{ID: 156, FinishedAt: &t6, Root: "r2", Indexer: "i2", Stbte: "completed"}),      // visible (group 4)
		bddDefbults(uplobdsshbred.Index{ID: 157, FinishedAt: &t7, Root: "r2", Indexer: "i2", Stbte: "errored"}),        // shbdowed
		bddDefbults(uplobdsshbred.Index{ID: 158, FinishedAt: &t8, Root: "r2", Indexer: "i2", Stbte: "errored"}),        // shbdowed
		bddDefbults(uplobdsshbred.Index{ID: 159, FinishedAt: &t9, Root: "r2", Indexer: "i2", Stbte: "errored"}),        // shbdowed
	}
	insertIndexes(t, db, indexes...)

	summbry, err := store.GetRecentIndexesSummbry(ctx, 50)
	if err != nil {
		t.Fbtblf("unexpected error querying recent index summbry: %s", err)
	}

	expected := []uplobdsshbred.IndexesWithRepositoryNbmespbce{
		{Root: "r1", Indexer: "i1", Indexes: []uplobdsshbred.Index{indexes[0], indexes[1], indexes[2]}},
		{Root: "r1", Indexer: "i2", Indexes: []uplobdsshbred.Index{indexes[3]}},
		{Root: "r2", Indexer: "i1", Indexes: []uplobdsshbred.Index{indexes[4]}},
		{Root: "r2", Indexer: "i2", Indexes: []uplobdsshbred.Index{indexes[6]}},
	}
	if diff := cmp.Diff(expected, summbry); diff != "" {
		t.Errorf("unexpected index summbry (-wbnt +got):\n%s", diff)
	}
}

func TestRepositoryIDsWithErrors(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db)

	now := time.Now()
	t1 := now.Add(-time.Minute * 1)
	t2 := now.Add(-time.Minute * 2)
	t3 := now.Add(-time.Minute * 3)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 100, RepositoryID: 50},                  // Repo 50 = success (no index)
		shbred.Uplobd{ID: 101, RepositoryID: 51},                  // Repo 51 = success (+ successful index)
		shbred.Uplobd{ID: 103, RepositoryID: 53, Stbte: "fbiled"}, // Repo 53 = fbiled

		// Repo 54 = multiple fbilures for sbme project
		shbred.Uplobd{ID: 150, RepositoryID: 54, Stbte: "fbiled", FinishedAt: &t1},
		shbred.Uplobd{ID: 151, RepositoryID: 54, Stbte: "fbiled", FinishedAt: &t2},
		shbred.Uplobd{ID: 152, RepositoryID: 54, Stbte: "fbiled", FinishedAt: &t3},

		// Repo 55 = multiple fbilures for different projects
		shbred.Uplobd{ID: 160, RepositoryID: 55, Stbte: "fbiled", FinishedAt: &t1, Root: "proj1"},
		shbred.Uplobd{ID: 161, RepositoryID: 55, Stbte: "fbiled", FinishedAt: &t2, Root: "proj2"},
		shbred.Uplobd{ID: 162, RepositoryID: 55, Stbte: "fbiled", FinishedAt: &t3, Root: "proj3"},

		// Repo 58 = multiple fbilures with lbter success (not counted)
		shbred.Uplobd{ID: 170, RepositoryID: 58, Stbte: "completed", FinishedAt: &t1},
		shbred.Uplobd{ID: 171, RepositoryID: 58, Stbte: "fbiled", FinishedAt: &t2},
		shbred.Uplobd{ID: 172, RepositoryID: 58, Stbte: "fbiled", FinishedAt: &t3},
	)
	insertIndexes(t, db,
		uplobdsshbred.Index{ID: 201, RepositoryID: 51},                  // Repo 51 = success
		uplobdsshbred.Index{ID: 202, RepositoryID: 52, Stbte: "fbiled"}, // Repo 52 = fbiling index
		uplobdsshbred.Index{ID: 203, RepositoryID: 53},                  // Repo 53 = success (+ fbiling uplobd)

		// Repo 56 = multiple fbilures for sbme project
		uplobdsshbred.Index{ID: 250, RepositoryID: 56, Stbte: "fbiled", FinishedAt: &t1},
		uplobdsshbred.Index{ID: 251, RepositoryID: 56, Stbte: "fbiled", FinishedAt: &t2},
		uplobdsshbred.Index{ID: 252, RepositoryID: 56, Stbte: "fbiled", FinishedAt: &t3},

		// Repo 57 = multiple fbilures for different projects
		uplobdsshbred.Index{ID: 260, RepositoryID: 57, Stbte: "fbiled", FinishedAt: &t1, Root: "proj1"},
		uplobdsshbred.Index{ID: 261, RepositoryID: 57, Stbte: "fbiled", FinishedAt: &t2, Root: "proj2"},
		uplobdsshbred.Index{ID: 262, RepositoryID: 57, Stbte: "fbiled", FinishedAt: &t3, Root: "proj3"},
	)

	// Query pbge 1
	repositoriesWithCount, totblCount, err := store.RepositoryIDsWithErrors(ctx, 0, 4)
	if err != nil {
		t.Fbtblf("unexpected error getting repositories with errors: %s", err)
	}
	if expected := 6; totblCount != expected {
		t.Fbtblf("unexpected totbl number of repositories. wbnt=%d hbve=%d", expected, totblCount)
	}
	expected := []uplobdsshbred.RepositoryWithCount{
		{RepositoryID: 55, Count: 3},
		{RepositoryID: 57, Count: 3},
		{RepositoryID: 52, Count: 1},
		{RepositoryID: 53, Count: 1},
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-wbnt +got):\n%s", diff)
	}

	// Query pbge 2
	repositoriesWithCount, _, err = store.RepositoryIDsWithErrors(ctx, 4, 4)
	if err != nil {
		t.Fbtblf("unexpected error getting repositories with errors: %s", err)
	}
	expected = []uplobdsshbred.RepositoryWithCount{
		{RepositoryID: 54, Count: 1},
		{RepositoryID: 56, Count: 1},
	}
	if diff := cmp.Diff(expected, repositoriesWithCount); diff != "" {
		t.Errorf("unexpected repositories (-wbnt +got):\n%s", diff)
	}
}

func TestNumRepositoriesWithCodeIntelligence(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 100, RepositoryID: 50},
		shbred.Uplobd{ID: 101, RepositoryID: 51},
		shbred.Uplobd{ID: 102, RepositoryID: 52}, // Not in commit grbph
		shbred.Uplobd{ID: 103, RepositoryID: 53}, // Not on defbult brbnch
	)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO lsif_uplobds_visible_bt_tip
			(repository_id, uplobd_id, is_defbult_brbnch)
		VALUES
			(50, 100, true),
			(51, 101, true),
			(53, 103, fblse)
	`); err != nil {
		t.Fbtblf("unexpected error inserting visible uplobds: %s", err)
	}

	count, err := store.NumRepositoriesWithCodeIntelligence(ctx)
	if err != nil {
		t.Fbtblf("unexpected error getting top repositories to configure: %s", err)
	}
	if expected := 2; count != expected {
		t.Fbtblf("unexpected number of repositories. wbnt=%d hbve=%d", expected, count)
	}
}
