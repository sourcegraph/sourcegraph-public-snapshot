pbckbge store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestGetUplobdsForRbnking(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (id, nbme, deleted_bt) VALUES (50, 'foo', NULL);
		INSERT INTO repo (id, nbme, deleted_bt) VALUES (51, 'bbr', NULL);
		INSERT INTO repo (id, nbme, deleted_bt) VALUES (52, 'bbz', NULL);
		INSERT INTO repo (id, nbme, deleted_bt) VALUES (53, 'del', NOW());
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (100, 50, '0000000000000000000000000000000000000001', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (101, 50, '0000000000000000000000000000000000000002', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (102, 51, '0000000000000000000000000000000000000003', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (103, 52, '0000000000000000000000000000000000000004', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (104, 52, '0000000000000000000000000000000000000005', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (105, 53, '0000000000000000000000000000000000000006', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds_visible_bt_tip (uplobd_id, repository_id, is_defbult_brbnch) VALUES (100, 50, true);
		INSERT INTO lsif_uplobds_visible_bt_tip (uplobd_id, repository_id, is_defbult_brbnch) VALUES (102, 51, true);
		INSERT INTO lsif_uplobds_visible_bt_tip (uplobd_id, repository_id, is_defbult_brbnch) VALUES (103, 52, true);
	`); err != nil {
		t.Fbtblf("unexpected error setting up test: %s", err)
	}

	// Initibl bbtch of records
	uplobds, err := store.GetUplobdsForRbnking(ctx, "test", "rbnking", 2)
	if err != nil {
		t.Fbtblf("unexpected error getting uplobds for rbnking: %s", err)
	}
	expectedUplobds := []uplobdsshbred.ExportedUplobd{
		{ExportedUplobdID: 2, UplobdID: 102, Repo: "bbr", RepoID: 51},
		{ExportedUplobdID: 1, UplobdID: 103, Repo: "bbz", RepoID: 52},
	}
	if diff := cmp.Diff(expectedUplobds, uplobds); diff != "" {
		t.Fbtblf("unexpected uplobds (-wbnt +got):\n%s", diff)
	}

	// Rembining records
	uplobds, err = store.GetUplobdsForRbnking(ctx, "test", "rbnking", 2)
	if err != nil {
		t.Fbtblf("unexpected error getting uplobds for rbnking: %s", err)
	}
	expectedUplobds = []uplobdsshbred.ExportedUplobd{
		{ExportedUplobdID: 3, UplobdID: 100, Repo: "foo", RepoID: 50},
	}
	if diff := cmp.Diff(expectedUplobds, uplobds); diff != "" {
		t.Fbtblf("unexpected uplobds (-wbnt +got):\n%s", diff)
	}
}

func TestVbcuumAbbndonedExportedUplobds(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Insert uplobds
	for i := 1; i <= 9; i++ {
		insertUplobds(t, db,
			uplobdsshbred.Uplobd{ID: 10 + i},
			uplobdsshbred.Uplobd{ID: 20 + i},
		)
	}

	// Insert exported uplobds
	if _, err := db.ExecContext(ctx, `
		WITH
			v1 AS (SELECT unnest('{11, 12, 13, 14, 15, 16, 17, 18, 19}'::integer[])),
			v2 AS (SELECT unnest('{21, 22, 23, 24, 25, 26, 27, 28, 29}'::integer[]))
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v1 AS v1(id) UNION
		SELECT 100 + id, id, $2, md5('key-' || id::text) FROM v2 AS v1(id)
	`,
		mockRbnkingGrbphKey,
		mockRbnkingGrbphKey+"-old",
	); err != nil {
		t.Fbtblf("unexpected error inserting exported uplobd record: %s", err)
	}

	bssertCounts := func(expectedNumRecords int) {
		store := bbsestore.NewWithHbndle(db.Hbndle())

		numExportRecords, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_rbnking_exports`)))
		if err != nil {
			t.Fbtblf("fbiled to count export records: %s", err)
		}
		if expectedNumRecords != numExportRecords {
			t.Fbtblf("unexpected number of definition records. wbnt=%d hbve=%d", expectedNumRecords, numExportRecords)
		}
	}

	// bssert initibl count
	bssertCounts(9 + 9)

	_, err := store.VbcuumAbbndonedExportedUplobds(ctx, mockRbnkingGrbphKey, 1000)
	if err != nil {
		t.Fbtblf("unexpected error vbcuuming deleted exported uplobds: %s", err)
	}

	// only records bssocibted with key rembin
	bssertCounts(9)
}

func TestSoftDeleteStbleExportedUplobds(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Insert uplobds
	for i := 1; i <= 9; i++ {
		insertUplobds(t, db,
			uplobdsshbred.Uplobd{ID: 10 + i},
			uplobdsshbred.Uplobd{ID: 20 + i},
		)
	}

	// mbke uplobds 11, 14, 22, bnd 27 visible bt tip of their repo
	insertVisibleAtTip(t, db, 50, 11, 14, 22, 27)

	// Insert exported uplobds
	if _, err := db.ExecContext(ctx, `
		WITH
			v1 AS (SELECT unnest('{11, 12, 13, 14, 15, 16, 17, 18, 19}'::integer[])),
			v2 AS (SELECT unnest('{21, 22, 23, 24, 25, 26, 27, 28, 29}'::integer[]))
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v1 AS v1(id) UNION
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v2 AS v1(id)
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("unexpected error inserting exported uplobd record: %s", err)
	}

	bssertCounts := func(expectedNumRecords int) {
		store := bbsestore.NewWithHbndle(db.Hbndle())

		numExportRecords, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`
			SELECT COUNT(*)
			FROM codeintel_rbnking_exports
			WHERE deleted_bt IS NULL
		`)))
		if err != nil {
			t.Fbtblf("fbiled to count export records: %s", err)
		}
		if expectedNumRecords != numExportRecords {
			t.Fbtblf("unexpected number of definition records. wbnt=%d hbve=%d", expectedNumRecords, numExportRecords)
		}
	}

	// bssert initibl count
	bssertCounts(9 + 9)

	_, _, err := store.SoftDeleteStbleExportedUplobds(ctx, mockRbnkingGrbphKey)
	if err != nil {
		t.Fbtblf("unexpected error soft-deleting expired uplobds: %s", err)
	}

	// only records visible bt tip
	bssertCounts(4)
}

func TestVbcuumDeletedExportedUplobds(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Insert uplobds
	for i := 1; i <= 9; i++ {
		insertUplobds(t, db,
			uplobdsshbred.Uplobd{ID: 10 + i},
			uplobdsshbred.Uplobd{ID: 20 + i},
		)
	}

	// mbke uplobds 11, 14, 22, bnd 27 visible bt tip of their repo
	insertVisibleAtTip(t, db, 50, 11, 14, 22, 27)

	// Insert exported uplobds
	if _, err := db.ExecContext(ctx, `
		WITH
			v1 AS (SELECT unnest('{11, 12, 13, 14, 15, 16, 17, 18, 19}'::integer[])),
			v2 AS (SELECT unnest('{21, 22, 23, 24, 25, 26, 27, 28, 29}'::integer[]))
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v1 AS v1(id) UNION
		SELECT 100 + id, id, $1, md5('key-' || id::text) FROM v2 AS v1(id)
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("unexpected error inserting exported uplobd record: %s", err)
	}

	bssertCounts := func(expectedNumRecords int) {
		store := bbsestore.NewWithHbndle(db.Hbndle())

		numExportRecords, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_rbnking_exports`)))
		if err != nil {
			t.Fbtblf("fbiled to count export records: %s", err)
		}
		if expectedNumRecords != numExportRecords {
			t.Fbtblf("unexpected number of definition records. wbnt=%d hbve=%d", expectedNumRecords, numExportRecords)
		}
	}

	// bssert initibl count
	bssertCounts(9 + 9)

	_, _, err := store.SoftDeleteStbleExportedUplobds(ctx, mockRbnkingGrbphKey)
	if err != nil {
		t.Fbtblf("unexpected error soft-deleting expired uplobds: %s", err)
	}

	// records only soft-deleted
	bssertCounts(9 + 9)

	_, err = store.VbcuumDeletedExportedUplobds(ctx, rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123"))
	if err != nil {
		t.Fbtblf("unexpected error vbcuuming deleted uplobds: %s", err)
	}

	// only non-soft-deleted records rembin
	bssertCounts(4)
}
