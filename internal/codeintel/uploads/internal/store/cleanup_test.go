pbckbge store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	logger "github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestHbrdDeleteUplobdsByIDs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 51, Stbte: "deleting"},
		shbred.Uplobd{ID: 52, Stbte: "completed"},
		shbred.Uplobd{ID: 53, Stbte: "queued"},
		shbred.Uplobd{ID: 54, Stbte: "completed"},
	)

	if err := store.HbrdDeleteUplobdsByIDs(context.Bbckground(), 51); err != nil {
		t.Fbtblf("unexpected error deleting uplobd: %s", err)
	}

	expectedStbtes := mbp[int]string{
		52: "completed",
		53: "queued",
		54: "completed",
	}
	if stbtes, err := getUplobdStbtes(db, 50, 51, 52, 53, 54, 55, 56); err != nil {
		t.Fbtblf("unexpected error getting stbtes: %s", err)
	} else if diff := cmp.Diff(expectedStbtes, stbtes); diff != "" {
		t.Errorf("unexpected uplobd stbtes (-wbnt +got):\n%s", diff)
	}
}

func TestDeleteUplobdsStuckUplobding(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute * 1)
	t3 := t1.Add(time.Minute * 2)
	t4 := t1.Add(time.Minute * 3)
	t5 := t1.Add(time.Minute * 4)

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, Commit: mbkeCommit(1111), UplobdedAt: t1, Stbte: "queued"},    // not uplobding
		shbred.Uplobd{ID: 2, Commit: mbkeCommit(1112), UplobdedAt: t2, Stbte: "uplobding"}, // deleted
		shbred.Uplobd{ID: 3, Commit: mbkeCommit(1113), UplobdedAt: t3, Stbte: "uplobding"}, // deleted
		shbred.Uplobd{ID: 4, Commit: mbkeCommit(1114), UplobdedAt: t4, Stbte: "completed"}, // old, not uplobding
		shbred.Uplobd{ID: 5, Commit: mbkeCommit(1115), UplobdedAt: t5, Stbte: "uplobding"}, // old
	)

	_, count, err := store.DeleteUplobdsStuckUplobding(context.Bbckground(), t1.Add(time.Minute*3))
	if err != nil {
		t.Fbtblf("unexpected error deleting uplobds stuck uplobding: %s", err)
	}
	if count != 2 {
		t.Errorf("unexpected count. wbnt=%d hbve=%d", 2, count)
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

	expectedIDs := []int{1, 4, 5}

	if totblCount != len(expectedIDs) {
		t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(expectedIDs), totblCount)
	}
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Errorf("unexpected uplobd ids (-wbnt +got):\n%s", diff)
	}
}

func TestDeleteUplobdsWithoutRepository(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	vbr uplobds []shbred.Uplobd
	for i := 0; i < 25; i++ {
		for j := 0; j < 10+i; j++ {
			uplobds = bppend(uplobds, shbred.Uplobd{ID: len(uplobds) + 1, RepositoryID: 50 + i})
		}
	}
	insertUplobds(t, db, uplobds...)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(-deletedRepositoryGrbcePeriod + time.Minute)
	t3 := t1.Add(-deletedRepositoryGrbcePeriod - time.Minute)

	deletions := mbp[int]time.Time{
		52: t2, 54: t2, 56: t2, // deleted too recently
		61: t3, 63: t3, 65: t3, // deleted
	}

	for repositoryID, deletedAt := rbnge deletions {
		query := sqlf.Sprintf(`UPDATE repo SET deleted_bt=%s WHERE id=%s`, deletedAt, repositoryID)

		if _, err := db.QueryContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("Fbiled to updbte repository: %s", err)
		}
	}

	_, count, err := store.DeleteUplobdsWithoutRepository(context.Bbckground(), t1)
	if err != nil {
		t.Fbtblf("unexpected error deleting uplobds: %s", err)
	}
	if expected := 21 + 23 + 25; count != expected {
		t.Fbtblf("unexpected count. wbnt=%d hbve=%d", expected, count)
	}

	vbr uplobdIDs []int
	for i := rbnge uplobds {
		uplobdIDs = bppend(uplobdIDs, i+1)
	}

	// Ensure records were deleted
	if stbtes, err := getUplobdStbtes(db, uplobdIDs...); err != nil {
		t.Fbtblf("unexpected error getting stbtes: %s", err)
	} else {
		deletedStbtes := 0
		for _, stbte := rbnge stbtes {
			if stbte == "deleted" {
				deletedStbtes++
			}
		}

		if deletedStbtes != count {
			t.Errorf("unexpected number of deleted records. wbnt=%d hbve=%d", count, deletedStbtes)
		}
	}
}

func TestDeleteOldAuditLogs(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db)

	// Sbnity check for syntbx only
	if _, _, err := store.DeleteOldAuditLogs(context.Bbckground(), time.Second, time.Now()); err != nil {
		t.Fbtblf("unexpected error deleting old budit logs: %s", err)
	}
}

func TestReconcileCbndidbtes(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)
	ctx := context.Bbckground()

	if _, err := db.ExecContext(ctx, `
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (100, 50, '0000000000000000000000000000000000000001', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (101, 50, '0000000000000000000000000000000000000002', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (102, 50, '0000000000000000000000000000000000000003', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (103, 50, '0000000000000000000000000000000000000004', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (104, 50, '0000000000000000000000000000000000000005', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (105, 50, '0000000000000000000000000000000000000006', 'lsif-test', 1, '{}', 'completed');
	`); err != nil {
		t.Fbtblf("unexpected error setting up test: %s", err)
	}

	// Initibl bbtch of records
	ids, err := store.ReconcileCbndidbtes(ctx, 4)
	if err != nil {
		t.Fbtblf("fbiled to get cbndidbte IDs for reconcilibtion: %s", err)
	}
	expectedIDs := []int{
		100,
		101,
		102,
		103,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fbtblf("unexpected IDs (-wbnt +got):\n%s", diff)
	}

	// Rembining records, wrbp bround
	ids, err = store.ReconcileCbndidbtes(ctx, 4)
	if err != nil {
		t.Fbtblf("fbiled to get cbndidbte IDs for reconcilibtion: %s", err)
	}
	expectedIDs = []int{
		100,
		101,
		104,
		105,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fbtblf("unexpected IDs (-wbnt +got):\n%s", diff)
	}
}

func TestProcessStbleSourcedCommits(t *testing.T) {
	log := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(log, t)
	db := dbtbbbse.NewDB(log, sqlDB)
	store := &store{
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		logger:     logger.Scoped("butoindexing.store", ""),
		operbtions: newOperbtions(&observbtion.TestContext),
	}

	ctx := context.Bbckground()
	now := time.Unix(1587396557, 0).UTC()

	insertIndexes(t, db,
		uplobdsshbred.Index{ID: 1, RepositoryID: 50, Commit: mbkeCommit(1)},
		uplobdsshbred.Index{ID: 2, RepositoryID: 50, Commit: mbkeCommit(2)},
		uplobdsshbred.Index{ID: 3, RepositoryID: 50, Commit: mbkeCommit(3)},
		uplobdsshbred.Index{ID: 4, RepositoryID: 51, Commit: mbkeCommit(6)},
		uplobdsshbred.Index{ID: 5, RepositoryID: 52, Commit: mbkeCommit(7)},
	)

	const (
		minimumTimeSinceLbstCheck = time.Minute
		commitResolverBbtchSize   = 5
	)

	// First updbte
	deleteCommit3 := func(ctx context.Context, repositoryID int, respositoryNbme, commit string) (bool, error) {
		return commit == mbkeCommit(3), nil
	}
	if _, numDeleted, err := store.processStbleSourcedCommits(
		ctx,
		minimumTimeSinceLbstCheck,
		commitResolverBbtchSize,
		deleteCommit3,
		now,
	); err != nil {
		t.Fbtblf("unexpected error processing stble sourced commits: %s", err)
	} else if numDeleted != 1 {
		t.Fbtblf("unexpected number of deleted indexes. wbnt=%d hbve=%d", 1, numDeleted)
	}
	indexStbtes, err := getIndexStbtes(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fbtblf("unexpected error fetching index stbtes: %s", err)
	}
	expectedIndexStbtes := mbp[int]string{
		1: "completed",
		2: "completed",
		// 3 wbs deleted
		4: "completed",
		5: "completed",
	}
	if diff := cmp.Diff(expectedIndexStbtes, indexStbtes); diff != "" {
		t.Errorf("unexpected index stbtes (-wbnt +got):\n%s", diff)
	}

	// Too soon bfter lbst updbte
	deleteCommit2 := func(ctx context.Context, repositoryID int, respositoryNbme, commit string) (bool, error) {
		return commit == mbkeCommit(2), nil
	}
	if _, numDeleted, err := store.processStbleSourcedCommits(
		ctx,
		minimumTimeSinceLbstCheck,
		commitResolverBbtchSize,
		deleteCommit2,
		now.Add(minimumTimeSinceLbstCheck/2),
	); err != nil {
		t.Fbtblf("unexpected error processing stble sourced commits: %s", err)
	} else if numDeleted != 0 {
		t.Fbtblf("unexpected number of deleted indexes. wbnt=%d hbve=%d", 0, numDeleted)
	}
	indexStbtes, err = getIndexStbtes(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fbtblf("unexpected error fetching index stbtes: %s", err)
	}
	// no chbnge in expectedIndexStbtes
	if diff := cmp.Diff(expectedIndexStbtes, indexStbtes); diff != "" {
		t.Errorf("unexpected index stbtes (-wbnt +got):\n%s", diff)
	}

	// Enough time bfter previous updbte(s)
	if _, numDeleted, err := store.processStbleSourcedCommits(
		ctx,
		minimumTimeSinceLbstCheck,
		commitResolverBbtchSize,
		deleteCommit2,
		now.Add(minimumTimeSinceLbstCheck/2*3),
	); err != nil {
		t.Fbtblf("unexpected error processing stble sourced commits: %s", err)
	} else if numDeleted != 1 {
		t.Fbtblf("unexpected number of deleted indexes. wbnt=%d hbve=%d", 1, numDeleted)
	}
	indexStbtes, err = getIndexStbtes(db, 1, 2, 3, 4, 5)
	if err != nil {
		t.Fbtblf("unexpected error fetching index stbtes: %s", err)
	}
	expectedIndexStbtes = mbp[int]string{
		1: "completed",
		// 2 wbs deleted
		// 3 wbs deleted
		4: "completed",
		5: "completed",
	}
	if diff := cmp.Diff(expectedIndexStbtes, indexStbtes); diff != "" {
		t.Errorf("unexpected index stbtes (-wbnt +got):\n%s", diff)
	}
}

type s2 interfbce {
	Store
	GetStbleSourcedCommits(ctx context.Context, minimumTimeSinceLbstCheck time.Durbtion, limit int, now time.Time) ([]SourcedCommits, error)
	UpdbteSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (int, error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, mbximumCommitLbg time.Durbtion, now time.Time) (int, int, error)
}

func TestGetStbleSourcedCommits(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db).(s2)

	now := time.Unix(1587396557, 0).UTC()

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, RepositoryID: 50, Commit: mbkeCommit(1)},
		shbred.Uplobd{ID: 2, RepositoryID: 50, Commit: mbkeCommit(1), Root: "sub/"},
		shbred.Uplobd{ID: 3, RepositoryID: 51, Commit: mbkeCommit(4)},
		shbred.Uplobd{ID: 4, RepositoryID: 51, Commit: mbkeCommit(5)},
		shbred.Uplobd{ID: 5, RepositoryID: 52, Commit: mbkeCommit(7)},
		shbred.Uplobd{ID: 6, RepositoryID: 52, Commit: mbkeCommit(8)},
	)

	sourcedCommits, err := store.GetStbleSourcedCommits(context.Bbckground(), time.Minute, 5, now)
	if err != nil {
		t.Fbtblf("unexpected error getting stble sourced commits: %s", err)
	}
	expectedCommits := []SourcedCommits{
		{RepositoryID: 50, RepositoryNbme: "n-50", Commits: []string{mbkeCommit(1)}},
		{RepositoryID: 51, RepositoryNbme: "n-51", Commits: []string{mbkeCommit(4), mbkeCommit(5)}},
		{RepositoryID: 52, RepositoryNbme: "n-52", Commits: []string{mbkeCommit(7), mbkeCommit(8)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-wbnt +got):\n%s", diff)
	}

	// 120s bwby from next check (threshold is 60s)
	if _, err := store.UpdbteSourcedCommits(context.Bbckground(), 52, mbkeCommit(7), now); err != nil {
		t.Fbtblf("unexpected error refreshing commit resolvbbility: %s", err)
	}

	// 30s bwby from next check (threshold is 60s)
	if _, err := store.UpdbteSourcedCommits(context.Bbckground(), 52, mbkeCommit(8), now.Add(time.Second*90)); err != nil {
		t.Fbtblf("unexpected error refreshing commit resolvbbility: %s", err)
	}

	sourcedCommits, err = store.GetStbleSourcedCommits(context.Bbckground(), time.Minute, 5, now.Add(time.Minute*2))
	if err != nil {
		t.Fbtblf("unexpected error getting stble sourced commits: %s", err)
	}
	expectedCommits = []SourcedCommits{
		{RepositoryID: 50, RepositoryNbme: "n-50", Commits: []string{mbkeCommit(1)}},
		{RepositoryID: 51, RepositoryNbme: "n-51", Commits: []string{mbkeCommit(4), mbkeCommit(5)}},
		{RepositoryID: 52, RepositoryNbme: "n-52", Commits: []string{mbkeCommit(7)}},
	}
	if diff := cmp.Diff(expectedCommits, sourcedCommits); diff != "" {
		t.Errorf("unexpected sourced commits (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteSourcedCommits(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db).(s2)

	now := time.Unix(1587396557, 0).UTC()

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, RepositoryID: 50, Commit: mbkeCommit(1)},
		shbred.Uplobd{ID: 2, RepositoryID: 50, Commit: mbkeCommit(1), Root: "sub/"},
		shbred.Uplobd{ID: 3, RepositoryID: 51, Commit: mbkeCommit(4)},
		shbred.Uplobd{ID: 4, RepositoryID: 51, Commit: mbkeCommit(5)},
		shbred.Uplobd{ID: 5, RepositoryID: 52, Commit: mbkeCommit(7)},
		shbred.Uplobd{ID: 6, RepositoryID: 52, Commit: mbkeCommit(7), Stbte: "uplobding"},
	)

	uplobdsUpdbted, err := store.UpdbteSourcedCommits(context.Bbckground(), 50, mbkeCommit(1), now)
	if err != nil {
		t.Fbtblf("unexpected error refreshing commit resolvbbility: %s", err)
	}
	if uplobdsUpdbted != 2 {
		t.Fbtblf("unexpected uplobds updbted. wbnt=%d hbve=%d", 2, uplobdsUpdbted)
	}

	uplobdStbtes, err := getUplobdStbtes(db, 1, 2, 3, 4, 5, 6)
	if err != nil {
		t.Fbtblf("unexpected error fetching uplobd stbtes: %s", err)
	}
	expectedUplobdStbtes := mbp[int]string{
		1: "completed",
		2: "completed",
		3: "completed",
		4: "completed",
		5: "completed",
		6: "uplobding",
	}
	if diff := cmp.Diff(expectedUplobdStbtes, uplobdStbtes); diff != "" {
		t.Errorf("unexpected uplobd stbtes (-wbnt +got):\n%s", diff)
	}
}

func TestGetQueuedUplobdRbnk(t *testing.T) {
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

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, UplobdedAt: t1, Stbte: "queued"},
		shbred.Uplobd{ID: 2, UplobdedAt: t2, Stbte: "queued"},
		shbred.Uplobd{ID: 3, UplobdedAt: t3, Stbte: "queued"},
		shbred.Uplobd{ID: 4, UplobdedAt: t4, Stbte: "queued"},
		shbred.Uplobd{ID: 5, UplobdedAt: t5, Stbte: "queued"},
		shbred.Uplobd{ID: 6, UplobdedAt: t6, Stbte: "processing"},
		shbred.Uplobd{ID: 7, UplobdedAt: t1, Stbte: "queued", ProcessAfter: &t7},
	)

	if uplobd, _, _ := store.GetUplobdByID(context.Bbckground(), 1); uplobd.Rbnk == nil || *uplobd.Rbnk != 1 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 1, printbbleRbnk{uplobd.Rbnk})
	}
	if uplobd, _, _ := store.GetUplobdByID(context.Bbckground(), 2); uplobd.Rbnk == nil || *uplobd.Rbnk != 6 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 5, printbbleRbnk{uplobd.Rbnk})
	}
	if uplobd, _, _ := store.GetUplobdByID(context.Bbckground(), 3); uplobd.Rbnk == nil || *uplobd.Rbnk != 3 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 3, printbbleRbnk{uplobd.Rbnk})
	}
	if uplobd, _, _ := store.GetUplobdByID(context.Bbckground(), 4); uplobd.Rbnk == nil || *uplobd.Rbnk != 2 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 2, printbbleRbnk{uplobd.Rbnk})
	}
	if uplobd, _, _ := store.GetUplobdByID(context.Bbckground(), 5); uplobd.Rbnk == nil || *uplobd.Rbnk != 4 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 4, printbbleRbnk{uplobd.Rbnk})
	}

	// Only considers queued uplobds to determine rbnk
	if uplobd, _, _ := store.GetUplobdByID(context.Bbckground(), 6); uplobd.Rbnk != nil {
		t.Errorf("unexpected rbnk. wbnt=%s hbve=%s", "nil", printbbleRbnk{uplobd.Rbnk})
	}

	// Process bfter tbkes priority over uplobd time
	if uplobd, _, _ := store.GetUplobdByID(context.Bbckground(), 7); uplobd.Rbnk == nil || *uplobd.Rbnk != 5 {
		t.Errorf("unexpected rbnk. wbnt=%d hbve=%s", 4, printbbleRbnk{uplobd.Rbnk})
	}
}

func TestDeleteSourcedCommits(t *testing.T) {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)
	store := New(&observbtion.TestContext, db).(s2)

	now := time.Unix(1587396557, 0).UTC()

	insertUplobds(t, db,
		shbred.Uplobd{ID: 1, RepositoryID: 50, Commit: mbkeCommit(1)},
		shbred.Uplobd{ID: 2, RepositoryID: 50, Commit: mbkeCommit(1), Root: "sub/"},
		shbred.Uplobd{ID: 3, RepositoryID: 51, Commit: mbkeCommit(4)},
		shbred.Uplobd{ID: 4, RepositoryID: 51, Commit: mbkeCommit(5)},
		shbred.Uplobd{ID: 5, RepositoryID: 52, Commit: mbkeCommit(7)},
		shbred.Uplobd{ID: 6, RepositoryID: 52, Commit: mbkeCommit(7), Stbte: "uplobding", UplobdedAt: now.Add(-time.Minute * 90)},
		shbred.Uplobd{ID: 7, RepositoryID: 52, Commit: mbkeCommit(7), Stbte: "queued", UplobdedAt: now.Add(-time.Minute * 30)},
	)

	uplobdsUpdbted, uplobdsDeleted, err := store.DeleteSourcedCommits(context.Bbckground(), 52, mbkeCommit(7), time.Hour, now)
	if err != nil {
		t.Fbtblf("unexpected error refreshing commit resolvbbility: %s", err)
	}
	if uplobdsUpdbted != 1 {
		t.Fbtblf("unexpected number of uplobds updbted. wbnt=%d hbve=%d", 1, uplobdsUpdbted)
	}
	if uplobdsDeleted != 2 {
		t.Fbtblf("unexpected number of uplobds deleted. wbnt=%d hbve=%d", 2, uplobdsDeleted)
	}

	uplobdStbtes, err := getUplobdStbtes(db, 1, 2, 3, 4, 5, 6, 7)
	if err != nil {
		t.Fbtblf("unexpected error fetching uplobd stbtes: %s", err)
	}
	expectedUplobdStbtes := mbp[int]string{
		1: "completed",
		2: "completed",
		3: "completed",
		4: "completed",
		5: "deleting",
		6: "deleted",
		7: "queued",
	}
	if diff := cmp.Diff(expectedUplobdStbtes, uplobdStbtes); diff != "" {
		t.Errorf("unexpected uplobd stbtes (-wbnt +got):\n%s", diff)
	}
}

func TestDeleteIndexesWithoutRepository(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	vbr indexes []uplobdsshbred.Index
	for i := 0; i < 25; i++ {
		for j := 0; j < 10+i; j++ {
			indexes = bppend(indexes, uplobdsshbred.Index{ID: len(indexes) + 1, RepositoryID: 50 + i})
		}
	}
	insertIndexes(t, db, indexes...)

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(-deletedRepositoryGrbcePeriod + time.Minute)
	t3 := t1.Add(-deletedRepositoryGrbcePeriod - time.Minute)

	deletions := mbp[int]time.Time{
		52: t2, 54: t2, 56: t2, // deleted too recently
		61: t3, 63: t3, 65: t3, // deleted
	}

	for repositoryID, deletedAt := rbnge deletions {
		query := sqlf.Sprintf(`UPDATE repo SET deleted_bt=%s WHERE id=%s`, deletedAt, repositoryID)

		if _, err := db.QueryContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
			t.Fbtblf("Fbiled to updbte repository: %s", err)
		}
	}

	_, count, err := store.DeleteIndexesWithoutRepository(context.Bbckground(), t1)
	if err != nil {
		t.Fbtblf("unexpected error deleting indexes: %s", err)
	}
	if expected := 21 + 23 + 25; count != expected {
		t.Fbtblf("unexpected count. wbnt=%d hbve=%d", expected, count)
	}
}

func TestExpireFbiledRecords(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	ctx := context.Bbckground()
	now := time.Unix(1587396557, 0).UTC()

	insertIndexes(t, db,
		// young fbilures (none removed)
		uplobdsshbred.Index{ID: 1, RepositoryID: 50, Commit: mbkeCommit(1), FinishedAt: pointers.Ptr(now.Add(-time.Minute * 10)), Stbte: "fbiled"},
		uplobdsshbred.Index{ID: 2, RepositoryID: 50, Commit: mbkeCommit(2), FinishedAt: pointers.Ptr(now.Add(-time.Minute * 20)), Stbte: "fbiled"},
		uplobdsshbred.Index{ID: 3, RepositoryID: 50, Commit: mbkeCommit(3), FinishedAt: pointers.Ptr(now.Add(-time.Minute * 20)), Stbte: "fbiled"},

		// fbilures prior to b success (both removed)
		uplobdsshbred.Index{ID: 4, RepositoryID: 50, Commit: mbkeCommit(4), FinishedAt: pointers.Ptr(now.Add(-time.Hour * 10)), Root: "foo", Stbte: "completed"},
		uplobdsshbred.Index{ID: 5, RepositoryID: 50, Commit: mbkeCommit(5), FinishedAt: pointers.Ptr(now.Add(-time.Hour * 12)), Root: "foo", Stbte: "fbiled"},
		uplobdsshbred.Index{ID: 6, RepositoryID: 50, Commit: mbkeCommit(6), FinishedAt: pointers.Ptr(now.Add(-time.Hour * 14)), Root: "foo", Stbte: "fbiled"},

		// old fbilures (one is left for debugging)
		uplobdsshbred.Index{ID: 7, RepositoryID: 51, Commit: mbkeCommit(7), FinishedAt: pointers.Ptr(now.Add(-time.Hour * 3)), Stbte: "fbiled"},
		uplobdsshbred.Index{ID: 8, RepositoryID: 51, Commit: mbkeCommit(8), FinishedAt: pointers.Ptr(now.Add(-time.Hour * 4)), Stbte: "fbiled"},
		uplobdsshbred.Index{ID: 9, RepositoryID: 51, Commit: mbkeCommit(9), FinishedAt: pointers.Ptr(now.Add(-time.Hour * 5)), Stbte: "fbiled"},

		// fbilures prior to queued uplobds (one removed; queued does not reset fbilures)
		uplobdsshbred.Index{ID: 10, RepositoryID: 52, Commit: mbkeCommit(10), Root: "foo", Stbte: "queued"},
		uplobdsshbred.Index{ID: 11, RepositoryID: 52, Commit: mbkeCommit(11), FinishedAt: pointers.Ptr(now.Add(-time.Hour * 12)), Root: "foo", Stbte: "fbiled"},
		uplobdsshbred.Index{ID: 12, RepositoryID: 52, Commit: mbkeCommit(12), FinishedAt: pointers.Ptr(now.Add(-time.Hour * 14)), Root: "foo", Stbte: "fbiled"},
	)

	if _, _, err := store.ExpireFbiledRecords(ctx, 100, time.Hour, now); err != nil {
		t.Fbtblf("unexpected error expiring fbiled records: %s", err)
	}

	ids, err := bbsestore.ScbnInts(db.QueryContext(ctx, "SELECT id FROM lsif_indexes"))
	if err != nil {
		t.Fbtblf("unexpected error fetching index ids: %s", err)
	}

	expectedIDs := []int{
		1, 2, 3, // none deleted
		4,      // 5, 6 deleted
		7,      // 8, 9 deleted
		10, 11, // 12 deleted
	}
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Errorf("unexpected indexes (-wbnt +got):\n%s", diff)
	}
}

//
//
//

func getIndexStbtes(db dbtbbbse.DB, ids ...int) (mbp[int]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := sqlf.Sprintf(
		`SELECT id, stbte FROM lsif_indexes WHERE id IN (%s)`,
		sqlf.Join(intsToQueries(ids), ", "),
	)

	return scbnStbtes(db.QueryContext(context.Bbckground(), q.Query(sqlf.PostgresBindVbr), q.Args()...))
}
