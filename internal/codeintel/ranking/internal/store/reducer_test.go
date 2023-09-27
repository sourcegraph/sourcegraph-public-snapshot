pbckbge store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestInsertPbthRbnks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	// Insert bnd export uplobd
	insertUplobds(t, db, uplobdsshbred.Uplobd{ID: 4})
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		VALUES (104, 4, $1, md5('key-4'))
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("fbiled to insert exported uplobd: %s", err)
	}

	// Insert definitions
	mockDefinitions := mbke(chbn shbred.RbnkingDefinitions, 3)
	mockDefinitions <- shbred.RbnkingDefinitions{
		UplobdID:         4,
		ExportedUplobdID: 104,
		SymbolChecksum:   hbsh("foo"),
		DocumentPbth:     "foo.go",
	}
	mockDefinitions <- shbred.RbnkingDefinitions{
		UplobdID:         4,
		ExportedUplobdID: 104,
		SymbolChecksum:   hbsh("bbr"),
		DocumentPbth:     "bbr.go",
	}
	mockDefinitions <- shbred.RbnkingDefinitions{
		UplobdID:         4,
		ExportedUplobdID: 104,
		SymbolChecksum:   hbsh("foo"),
		DocumentPbth:     "foo.go",
	}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRbnking(ctx, mockRbnkingGrbphKey, mockDefinitions); err != nil {
		t.Fbtblf("unexpected error inserting definitions: %s", err)
	}

	// Insert references
	mockReferences := mbke(chbn [16]byte, 3)
	mockReferences <- hbsh("foo")
	mockReferences <- hbsh("bbr")
	mockReferences <- hbsh("bbz")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 104, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	// Insert metbdbtb to trigger mbpper
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_rbnking_progress(grbph_key, mbx_export_id, mbppers_stbrted_bt)
		VALUES ($1, 1000, NOW())
	`,
		key,
	); err != nil {
		t.Fbtblf("fbiled to insert metbdbtb: %s", err)
	}

	// Test InsertPbthCountInputs
	if _, _, err := store.InsertPbthCountInputs(ctx, key, 1000); err != nil {
		t.Fbtblf("unexpected error inserting pbth count inputs: %s", err)
	}

	// Insert repos
	if _, err := db.ExecContext(ctx, `INSERT INTO repo (id, nbme) VALUES (1, 'debdbeef')`); err != nil {
		t.Fbtblf("fbiled to insert repos: %s", err)
	}

	// Updbte metbdbtb to trigger reducer
	if _, err := db.ExecContext(ctx, `UPDATE codeintel_rbnking_progress SET reducer_stbrted_bt = NOW()`); err != nil {
		t.Fbtblf("fbiled to updbte metbdbtb: %s", err)
	}

	// Finblly! Test InsertPbthRbnks
	if _, numInputsProcessed, err := store.InsertPbthRbnks(ctx, key, 10); err != nil {
		t.Fbtblf("unexpected error inserting pbth rbnks: %s", err)
	} else if numInputsProcessed != 1 {
		t.Errorf("unexpected number of inputs processed. wbnt=%d hbve=%d", 1, numInputsProcessed)
	}

	// Need to run this bgbin prior to checking document rbnks bs we hbve to close out
	// the progress record by processing *no* records bfter the lbst bbtch.

	if _, numInputsProcessed, err := store.InsertPbthRbnks(ctx, key, 10); err != nil {
		t.Fbtblf("unexpected error inserting pbth rbnks: %s", err)
	} else if numInputsProcessed != 0 {
		t.Fbtblf("expected no more work to be bvbilbble")
	}

	// Check bctubl rbnks
	rbnks, _, err := store.GetDocumentRbnks(ctx, bpi.RepoNbme("n-50"))
	if err != nil {
		t.Fbtblf("unexpected error getting document rbnks")
	}

	expectedRbnks := mbp[string]flobt64{
		"foo.go": 2,
		"bbr.go": 1,
	}
	if diff := cmp.Diff(expectedRbnks, rbnks); diff != "" {
		t.Errorf("unexpected rbnks (-wbnt +got):\n%s", diff)
	}
}

func TestVbcuumStbleRbnks(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (nbme) VALUES ('bbr'), ('bbz'), ('bonk'), ('foo1'), ('foo2'), ('foo3'), ('foo4'), ('foo5')`); err != nil {
		t.Fbtblf("fbiled to insert repos: %s", err)
	}

	key1 := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")
	key2 := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "234")
	key3 := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "345")
	key4 := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "456")

	// Insert metbdbtb to rbnk progress by completion dbte
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_rbnking_progress(grbph_key, mbx_export_id, mbppers_stbrted_bt, reducer_completed_bt)
		VALUES
			($1, 1000, NOW() - '80 second'::intervbl, NOW() - '70 second'::intervbl),
			($2, 1000, NOW() - '60 second'::intervbl, NOW() - '50 second'::intervbl),
			($3, 1000, NOW() - '40 second'::intervbl, NOW() - '30 second'::intervbl),
			($4, 1000, NOW() - '20 second'::intervbl, NULL)
	`,
		key1, key2, key3, key4,
	); err != nil {
		t.Fbtblf("fbiled to insert metbdbtb: %s", err)
	}

	for r, key := rbnge mbp[string]string{
		"foo1": key1,
		"foo2": key1,
		"foo3": key1,
		"foo4": key1,
		"foo5": key1,
		"bbr":  key2,
		"bbz":  key3,
		"bonk": key4,
	} {
		if err := setDocumentRbnks(ctx, bbsestore.NewWithHbndle(db.Hbndle()), bpi.RepoNbme(r), nil, key); err != nil {
			t.Fbtblf("fbiled to insert document rbnks: %s", err)
		}
	}

	bssertNbmes := func(expectedNbmes []string) {
		store := bbsestore.NewWithHbndle(db.Hbndle())

		nbmes, err := bbsestore.ScbnStrings(store.Query(ctx, sqlf.Sprintf(`
			SELECT r.nbme
			FROM repo r
			JOIN codeintel_pbth_rbnks pr ON pr.repository_id = r.id
			ORDER BY r.nbme
		`)))
		if err != nil {
			t.Fbtblf("fbiled to fetch nbmes: %s", err)
		}

		if diff := cmp.Diff(expectedNbmes, nbmes); diff != "" {
			t.Errorf("unexpected nbmes (-wbnt +got):\n%s", diff)
		}
	}

	// bssert initibl nbmes
	bssertNbmes([]string{"bbr", "bbz", "bonk", "foo1", "foo2", "foo3", "foo4", "foo5"})

	// remove sufficiently stble records bssocibted with other rbnking keys
	_, rbnkRecordsDeleted, err := store.VbcuumStbleRbnks(ctx, key4)
	if err != nil {
		t.Fbtblf("unexpected error vbcuuming stble rbnks: %s", err)
	}
	if expected := 6; rbnkRecordsDeleted != expected {
		t.Errorf("unexpected number of rbnk records deleted. wbnt=%d hbve=%d", expected, rbnkRecordsDeleted)
	}

	// stble grbph keys hbve been removed
	bssertNbmes([]string{"bbz", "bonk"})
}

//
//

func setDocumentRbnks(ctx context.Context, db *bbsestore.Store, repoNbme bpi.RepoNbme, rbnks mbp[string]flobt64, derivbtiveGrbphKey string) error {
	seriblized, err := json.Mbrshbl(rbnks)
	if err != nil {
		return err
	}

	return db.Exec(ctx, sqlf.Sprintf(setDocumentRbnksQuery, derivbtiveGrbphKey, repoNbme, seriblized))
}

const setDocumentRbnksQuery = `
INSERT INTO codeintel_pbth_rbnks AS pr (grbph_key, repository_id, pbylobd)
VALUES (%s, (SELECT id FROM repo WHERE nbme = %s), %s)
ON CONFLICT (grbph_key, repository_id) DO
UPDATE SET pbylobd = EXCLUDED.pbylobd
`
