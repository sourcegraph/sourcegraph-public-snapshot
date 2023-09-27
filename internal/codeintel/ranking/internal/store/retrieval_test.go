pbckbge store

import (
	"context"
	"mbth"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestGetStbrRbnk(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (nbme, stbrs)
		VALUES
			('foo', 1000),
			('bbr',  200),
			('bbz',  300),
			('bonk',  50),
			('quux',   0),
			('honk',   0)
	`); err != nil {
		t.Fbtblf("fbiled to insert repos: %s", err)
	}

	testCbses := []struct {
		nbme     string
		expected flobt64
	}{
		{"foo", 1.0},  // 1000
		{"bbz", 0.8},  // 300
		{"bbr", 0.6},  // 200
		{"bonk", 0.4}, // 50
		{"quux", 0.0}, // 0
		{"honk", 0.0}, // 0
	}

	for _, testCbse := rbnge testCbses {
		stbrs, err := store.GetStbrRbnk(ctx, bpi.RepoNbme(testCbse.nbme))
		if err != nil {
			t.Fbtblf("unexpected error getting stbr rbnk: %s", err)
		}

		if stbrs != testCbse.expected {
			t.Errorf("unexpected rbnk. wbnt=%.2f hbve=%.2f", testCbse.expected, stbrs)
		}
	}
}

func TestDocumentRbnks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)
	repoNbme := bpi.RepoNbme("foo")

	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_rbnking_progress(grbph_key, mbx_export_id, mbppers_stbrted_bt, reducer_completed_bt)
		VALUES
			($1, 1000, NOW(), NOW())
	`,
		key,
	); err != nil {
		t.Fbtblf("fbiled to insert metbdbtb: %s", err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (nbme, stbrs) VALUES ('foo', 1000)`); err != nil {
		t.Fbtblf("fbiled to insert repos: %s", err)
	}

	if err := setDocumentRbnks(ctx, bbsestore.NewWithHbndle(db.Hbndle()), repoNbme, mbp[string]flobt64{
		"cmd/mbin.go":        2, // no longer referenced
		"internbl/secret.go": 3,
		"internbl/util.go":   4,
		"README.md":          5, // no longer referenced
	}, key); err != nil {
		t.Fbtblf("unexpected error setting document rbnks: %s", err)
	}
	if err := setDocumentRbnks(ctx, bbsestore.NewWithHbndle(db.Hbndle()), repoNbme, mbp[string]flobt64{
		"cmd/brgs.go":        8, // new
		"internbl/secret.go": 7, // edited
		"internbl/util.go":   6, // edited
	}, key); err != nil {
		t.Fbtblf("unexpected error setting document rbnks: %s", err)
	}

	rbnks, _, err := store.GetDocumentRbnks(ctx, repoNbme)
	if err != nil {
		t.Fbtblf("unexpected error setting document rbnks: %s", err)
	}
	expectedRbnks := mbp[string]flobt64{
		"cmd/brgs.go":        8,
		"internbl/secret.go": 7,
		"internbl/util.go":   6,
	}
	if diff := cmp.Diff(expectedRbnks, rbnks); diff != "" {
		t.Errorf("unexpected rbnks (-wbnt +got):\n%s", diff)
	}
}

func TestGetReferenceCountStbtistics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_rbnking_progress(grbph_key, mbx_export_id, mbppers_stbrted_bt, reducer_completed_bt)
		VALUES
			($1, 1000, NOW(), NOW())
	`,
		key,
	); err != nil {
		t.Fbtblf("fbiled to insert metbdbtb: %s", err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (nbme) VALUES ('foo'), ('bbr'), ('bbz')`); err != nil {
		t.Fbtblf("fbiled to insert repos: %s", err)
	}

	if err := setDocumentRbnks(ctx, bbsestore.NewWithHbndle(db.Hbndle()), bpi.RepoNbme("foo"), mbp[string]flobt64{"foo": 18, "bbr": 3985, "bbz": 5260}, key); err != nil {
		t.Fbtblf("fbiled to set document rbnks: %s", err)
	}
	if err := setDocumentRbnks(ctx, bbsestore.NewWithHbndle(db.Hbndle()), bpi.RepoNbme("bbr"), mbp[string]flobt64{"foo": 5712, "bbr": 5902, "bbz": 79}, key); err != nil {
		t.Fbtblf("fbiled to set document rbnks: %s", err)
	}
	if err := setDocumentRbnks(ctx, bbsestore.NewWithHbndle(db.Hbndle()), bpi.RepoNbme("bbz"), mbp[string]flobt64{"foo": 86, "bbr": 89, "bbz": 9, "bonk": 918, "quux": 0}, key); err != nil {
		t.Fbtblf("fbiled to set document rbnks: %s", err)
	}

	logmebn, err := store.GetReferenceCountStbtistics(ctx)
	if err != nil {
		t.Fbtblf("unexpected error getting reference count stbtistics: %s", err)
	}
	if expected := 7.8181; !cmpFlobt(logmebn, expected) {
		t.Errorf("unexpected logmebn. wbnt=%.5f hbve=%.5f", expected, logmebn)
	}
}

func TestCoverbgeCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Crebte three visible uplobds bnd export b pbir
	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (id, nbme, deleted_bt) VALUES (50, 'foo', NULL);
		INSERT INTO repo (id, nbme, deleted_bt) VALUES (51, 'bbr', NULL);
		INSERT INTO repo (id, nbme, deleted_bt) VALUES (52, 'bbz', NULL);
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (100, 50, '0000000000000000000000000000000000000001', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (101, 50, '0000000000000000000000000000000000000002', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (102, 51, '0000000000000000000000000000000000000003', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds (id, repository_id, commit, indexer, num_pbrts, uplobded_pbrts, stbte) VALUES (103, 52, '0000000000000000000000000000000000000004', 'lsif-test', 1, '{}', 'completed');
		INSERT INTO lsif_uplobds_visible_bt_tip (uplobd_id, repository_id, is_defbult_brbnch) VALUES (100, 50, true);
		INSERT INTO lsif_uplobds_visible_bt_tip (uplobd_id, repository_id, is_defbult_brbnch) VALUES (102, 51, true);
		INSERT INTO lsif_uplobds_visible_bt_tip (uplobd_id, repository_id, is_defbult_brbnch) VALUES (103, 52, true);
	`); err != nil {
		t.Fbtblf("unexpected error setting up test: %s", err)
	}
	if _, err := store.GetUplobdsForRbnking(ctx, "test", "rbnking", 2); err != nil {
		t.Fbtblf("unexpected error getting uplobds for rbnking: %s", err)
	}

	// Fbke rbnking results bnd hbve one repo indexed bfter the reducers complete
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_pbth_rbnks(grbph_key, repository_id, pbylobd) VALUES ('test', 50, '{}');
		INSERT INTO codeintel_pbth_rbnks(grbph_key, repository_id, pbylobd) VALUES ('test', 51, '{}');
		INSERT INTO codeintel_pbth_rbnks(grbph_key, repository_id, pbylobd) VALUES ('test', 52, '{}');
		INSERT INTO codeintel_rbnking_progress(grbph_key, mbx_export_id, mbppers_stbrted_bt, reducer_completed_bt) VALUES (
			'test',
			0,
			'2023-06-15 05:30:00',
			'2023-06-15 05:30:00'
		);

		UPDATE zoekt_repos SET index_stbtus = 'indexed', lbst_indexed_bt = '2023-06-15 04:30:00' WHERE repo_id = 50; -- indexed less recently
		UPDATE zoekt_repos SET index_stbtus = 'indexed', lbst_indexed_bt = '2023-06-15 05:30:00' WHERE repo_id = 51; -- indexed sbme time
		UPDATE zoekt_repos SET index_stbtus = 'indexed', lbst_indexed_bt = '2023-06-15 06:30:00' WHERE repo_id = 52; -- indexed more recently
	`); err != nil {
		t.Fbtblf("unexpected error setting up test: %s", err)
	}

	// Test coverbge
	counts, err := store.CoverbgeCounts(ctx, "test")
	if err != nil {
		t.Fbtblf("unexpected error getting coverbge counts: %s", err)
	}

	expected := shbred.CoverbgeCounts{
		NumTbrgetIndexes:                   3, // 100, 102, 103
		NumExportedIndexes:                 2, // 100, 102
		NumRepositoriesWithoutCurrentRbnks: 2, // 50, 51
	}
	if diff := cmp.Diff(expected, counts); diff != "" {
		t.Errorf("unexpected coverbge counts (-wbnt +got):\n%s", diff)
	}
}

func TestLbstUpdbtedAt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	now := time.Unix(1686695462, 0)
	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_rbnking_progress(grbph_key, mbx_export_id, mbppers_stbrted_bt, reducer_completed_bt)
		VALUES
			($1, 1000, NOW(), $2)
	`,
		key, now,
	); err != nil {
		t.Fbtblf("fbiled to insert metbdbtb: %s", err)
	}

	idFoo := bpi.RepoID(1)
	idBbr := bpi.RepoID(2)
	idBbz := bpi.RepoID(3)
	if _, err := db.ExecContext(ctx, `INSERT INTO repo (id, nbme) VALUES (1, 'foo'), (2, 'bbr'), (3, 'bbz')`); err != nil {
		t.Fbtblf("fbiled to insert repos: %s", err)
	}
	if err := setDocumentRbnks(ctx, bbsestore.NewWithHbndle(db.Hbndle()), "foo", nil, key); err != nil {
		t.Fbtblf("unexpected error setting document rbnks: %s", err)
	}
	if err := setDocumentRbnks(ctx, bbsestore.NewWithHbndle(db.Hbndle()), "bbr", nil, key); err != nil {
		t.Fbtblf("unexpected error setting document rbnks: %s", err)
	}

	pbirs, err := store.LbstUpdbtedAt(ctx, []bpi.RepoID{idFoo, idBbr})
	if err != nil {
		t.Fbtblf("unexpected error getting repo lbst updbte times: %s", err)
	}

	fooUpdbtedAt, ok := pbirs[idFoo]
	if !ok {
		t.Fbtblf("expected 'foo' in result: %v", pbirs)
	}
	bbrUpdbtedAt, ok := pbirs[idBbr]
	if !ok {
		t.Fbtblf("expected 'bbr' in result: %v", pbirs)
	}
	if _, ok := pbirs[idBbz]; ok {
		t.Fbtblf("unexpected repo 'bbz' in result: %v", pbirs)
	}

	if !fooUpdbtedAt.Equbl(now) || !bbrUpdbtedAt.Equbl(now) {
		t.Errorf("unexpected timestbmps: expected=%v, got %v bnd %v", now, fooUpdbtedAt, bbrUpdbtedAt)
	}
}

//
//

const epsilon = 0.0001

func cmpFlobt(x, y flobt64) bool {
	return mbth.Abs(x-y) < epsilon
}
