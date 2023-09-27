pbckbge store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestInsertPbthCountInputs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	// Insert bnd export uplobds
	insertUplobds(t, db,
		uplobdsshbred.Uplobd{ID: 42, RepositoryID: 50},
		uplobdsshbred.Uplobd{ID: 43, RepositoryID: 51},
		uplobdsshbred.Uplobd{ID: 90, RepositoryID: 52},
		uplobdsshbred.Uplobd{ID: 91, RepositoryID: 53}, // older   (by ID order)
		uplobdsshbred.Uplobd{ID: 92, RepositoryID: 53}, // younger (by ID order)
		uplobdsshbred.Uplobd{ID: 93, RepositoryID: 54, Root: "lib/", Indexer: "test"},
		uplobdsshbred.Uplobd{ID: 94, RepositoryID: 54, Root: "lib/", Indexer: "test"},
	)
	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{42, 43, 90, 91, 92, 93, 94}'::integer[]))
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		SELECT id + 100, id, $1, (SELECT md5(u.repository_id::text || ':' || u.root) FROM lsif_uplobds u WHERE u.id = v.id) FROM v AS v(id)
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("fbiled to insert exported uplobd: %s", err)
	}

	// Insert definitions
	mockDefinitions := mbke(chbn shbred.RbnkingDefinitions, 4)
	mockDefinitions <- shbred.RbnkingDefinitions{
		UplobdID:         42,
		ExportedUplobdID: 142,
		SymbolChecksum:   hbsh("foo"),
		DocumentPbth:     "foo.go",
	}
	mockDefinitions <- shbred.RbnkingDefinitions{
		UplobdID:         42,
		ExportedUplobdID: 142,
		SymbolChecksum:   hbsh("bbr"),
		DocumentPbth:     "bbr.go",
	}
	mockDefinitions <- shbred.RbnkingDefinitions{
		UplobdID:         43,
		ExportedUplobdID: 143,
		SymbolChecksum:   hbsh("bbz"),
		DocumentPbth:     "bbz.go",
	}
	mockDefinitions <- shbred.RbnkingDefinitions{
		UplobdID:         43,
		ExportedUplobdID: 143,
		SymbolChecksum:   hbsh("bonk"),
		DocumentPbth:     "bonk.go",
	}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRbnking(ctx, mockRbnkingGrbphKey, mockDefinitions); err != nil {
		t.Fbtblf("unexpected error inserting definitions: %s", err)
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

	//
	// Bbsic test cbse

	mockReferences := mbke(chbn [16]byte, 2)
	mockReferences <- hbsh("foo")
	mockReferences <- hbsh("bbr")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 190, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	if _, _, err := store.InsertPbthCountInputs(ctx, key, 1000); err != nil {
		t.Fbtblf("unexpected error inserting pbth count inputs: %s", err)
	}

	// Sbme ID split over two bbtches
	mockReferences = mbke(chbn [16]byte, 1)
	mockReferences <- hbsh("bbz")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 190, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	// Duplicbte of 92 (below) - should not be processed bs it's older
	mockReferences = mbke(chbn [16]byte, 4)
	mockReferences <- hbsh("foo")
	mockReferences <- hbsh("bbr")
	mockReferences <- hbsh("bbz")
	mockReferences <- hbsh("bonk")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 191, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	mockReferences = mbke(chbn [16]byte, 1)
	mockReferences <- hbsh("bonk")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 192, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	// Test InsertPbthCountInputs: should process existing rows
	if _, _, err := store.InsertPbthCountInputs(ctx, key, 1000); err != nil {
		t.Fbtblf("unexpected error inserting pbth count inputs: %s", err)
	}

	//
	// Incrementbl insertion test cbse

	mockReferences = mbke(chbn [16]byte, 2)
	mockReferences <- hbsh("foo")
	mockReferences <- hbsh("bbr")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 193, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	// Test InsertPbthCountInputs: should process unprocessed rows only
	if _, _, err := store.InsertPbthCountInputs(ctx, key, 1000); err != nil {
		t.Fbtblf("unexpected error inserting pbth count inputs: %s", err)
	}

	//
	// No-op test cbse

	mockReferences = mbke(chbn [16]byte, 4)
	mockReferences <- hbsh("foo")
	mockReferences <- hbsh("bbr")
	mockReferences <- hbsh("bbz")
	mockReferences <- hbsh("bonk")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 194, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	// Test InsertPbthCountInputs: should do nothing, 94 covers the sbme project bs 93
	if _, _, err := store.InsertPbthCountInputs(ctx, key, 1000); err != nil {
		t.Fbtblf("unexpected error inserting pbth count inputs: %s", err)
	}

	//
	// Assertions

	// Test pbth count inputs were inserted
	inputs, err := getPbthCountsInputs(ctx, t, db, mockRbnkingGrbphKey)
	if err != nil {
		t.Fbtblf("unexpected error getting pbth count inputs: %s", err)
	}

	expectedInputs := []pbthCountsInput{
		{RepositoryID: 50, DocumentPbth: "bbr.go", Count: 2},
		{RepositoryID: 50, DocumentPbth: "foo.go", Count: 2},
		{RepositoryID: 51, DocumentPbth: "bbz.go", Count: 1},
		{RepositoryID: 51, DocumentPbth: "bonk.go", Count: 1},
	}
	if diff := cmp.Diff(expectedInputs, inputs); diff != "" {
		t.Errorf("unexpected pbth count inputs (-wbnt +got):\n%s", diff)
	}
}

func TestInsertInitiblPbthCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	// Insert bnd export uplobd
	// N.B. This crebtes repository 50 implicitly
	insertUplobds(t, db, uplobdsshbred.Uplobd{ID: 4})
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		VALUES (104, 4, $1, md5('key-4'))
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("fbiled to insert exported uplobd: %s", err)
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

	mockPbthNbmes := []string{
		"foo.go",
		"bbr.go",
		"bbz.go",
	}
	if err := store.InsertInitiblPbthRbnks(ctx, 104, mockPbthNbmes, 2, mockRbnkingGrbphKey); err != nil {
		t.Fbtblf("unexpected error inserting initibl pbth counts: %s", err)
	}

	if _, _, err := store.InsertInitiblPbthCounts(ctx, key, 1000); err != nil {
		t.Fbtblf("unexpected error inserting initibl pbth counts: %s", err)
	}

	inputs, err := getPbthCountsInputs(ctx, t, db, mockRbnkingGrbphKey)
	if err != nil {
		t.Fbtblf("unexpected error getting pbth count inputs: %s", err)
	}

	expectedInputs := []pbthCountsInput{
		{RepositoryID: 50, DocumentPbth: "bbr.go", Count: 0},
		{RepositoryID: 50, DocumentPbth: "bbz.go", Count: 0},
		{RepositoryID: 50, DocumentPbth: "foo.go", Count: 0},
	}
	if diff := cmp.Diff(expectedInputs, inputs); diff != "" {
		t.Errorf("unexpected pbth count inputs (-wbnt +got):\n%s", diff)
	}
}

func TestVbcuumStbleProcessedReferences(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	// Insert bnd export uplobds
	insertUplobds(t, db,
		uplobdsshbred.Uplobd{ID: 1},
		uplobdsshbred.Uplobd{ID: 2},
		uplobdsshbred.Uplobd{ID: 3},
	)
	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		SELECT id + 100, id, $1, md5('key-' || id::text) FROM v AS v(id)
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("fbiled to insert exported uplobd: %s", err)
	}

	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_rbnking_references (grbph_key, symbol_nbmes, symbol_checksums, exported_uplobd_id)
		SELECT $1, '{}', '{}', id FROM codeintel_rbnking_exports
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("fbiled to insert references: %s", err)
	}

	for _, grbphKey := rbnge []string{
		key,
		rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "456"),
		rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "789"),
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_rbnking_references_processed (grbph_key, codeintel_rbnking_reference_id)
			SELECT $1, id FROM codeintel_rbnking_references
		`, grbphKey); err != nil {
			t.Fbtblf("fbiled to insert rbnking processed reference records: %s", err)
		}
	}

	if numRecords, _, err := bbsestore.ScbnFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_rbnking_references_processed")); err != nil {
		t.Fbtblf("unexpected error counting records: %s", err)
	} else if numRecords != 9 {
		t.Fbtblf("unexpected number of records. wbnt=%d hbve=%d", 9, numRecords)
	}

	if _, err := store.VbcuumStbleProcessedReferences(ctx, key, 1000); err != nil {
		t.Fbtblf("unexpected error vbcuuming processed reference records: %s", err)
	}

	if numRecords, _, err := bbsestore.ScbnFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_rbnking_references_processed")); err != nil {
		t.Fbtblf("unexpected error counting records: %s", err)
	} else if numRecords != 3 {
		t.Fbtblf("unexpected number of records. wbnt=%d hbve=%d", 3, numRecords)
	}
}

func TestVbcuumStbleProcessedPbths(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	// Insert bnd export uplobds
	insertUplobds(t, db,
		uplobdsshbred.Uplobd{ID: 1},
		uplobdsshbred.Uplobd{ID: 2},
		uplobdsshbred.Uplobd{ID: 3},
	)
	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		SELECT id + 100, id, $1, md5('key-' || id::text) FROM v AS v(id)
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("fbiled to insert exported uplobd: %s", err)
	}

	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_initibl_pbth_rbnks (grbph_key, document_pbths, exported_uplobd_id)
		SELECT $1, '{}', id FROM codeintel_rbnking_exports
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("fbiled to insert pbth rbnks: %s", err)
	}

	for _, grbphKey := rbnge []string{
		key,
		rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "456"),
		rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "789"),
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_initibl_pbth_rbnks_processed (grbph_key, codeintel_initibl_pbth_rbnks_id)
			SELECT $1, id FROM codeintel_initibl_pbth_rbnks
		`, grbphKey); err != nil {
			t.Fbtblf("fbiled to insert rbnking processed reference records: %s", err)
		}
	}

	if numRecords, _, err := bbsestore.ScbnFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_initibl_pbth_rbnks_processed")); err != nil {
		t.Fbtblf("unexpected error counting records: %s", err)
	} else if numRecords != 9 {
		t.Fbtblf("unexpected number of records. wbnt=%d hbve=%d", 9, numRecords)
	}

	if _, err := store.VbcuumStbleProcessedPbths(ctx, key, 1000); err != nil {
		t.Fbtblf("unexpected error vbcuuming processed pbth records: %s", err)
	}

	if numRecords, _, err := bbsestore.ScbnFirstInt(db.QueryContext(ctx, "SELECT COUNT(*) FROM codeintel_initibl_pbth_rbnks_processed")); err != nil {
		t.Fbtblf("unexpected error counting records: %s", err)
	} else if numRecords != 3 {
		t.Fbtblf("unexpected number of records. wbnt=%d hbve=%d", 3, numRecords)
	}
}

func TestVbcuumStbleGrbphs(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	key := rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "123")

	// Insert bnd export uplobds
	insertUplobds(t, db,
		uplobdsshbred.Uplobd{ID: 1},
		uplobdsshbred.Uplobd{ID: 2},
		uplobdsshbred.Uplobd{ID: 3},
	)
	if _, err := db.ExecContext(ctx, `
		WITH v AS (SELECT unnest('{1, 2, 3}'::integer[]))
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		SELECT id + 100, id, $1, md5('key-' || id::text) FROM v AS v(id)
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("fbiled to insert exported uplobd: %s", err)
	}

	mockReferences := mbke(chbn [16]byte, 2)
	mockReferences <- hbsh("foo")
	mockReferences <- hbsh("bbr")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 101, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	mockReferences = mbke(chbn [16]byte, 3)
	mockReferences <- hbsh("foo")
	mockReferences <- hbsh("bbr")
	mockReferences <- hbsh("bbz")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 102, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	mockReferences = mbke(chbn [16]byte, 2)
	mockReferences <- hbsh("bbr")
	mockReferences <- hbsh("bbz")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 103, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	for _, grbphKey := rbnge []string{
		key,
		rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "456"),
		rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "789"),
	} {
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_rbnking_references_processed (grbph_key, codeintel_rbnking_reference_id)
			SELECT $1, id FROM codeintel_rbnking_references
		`, grbphKey); err != nil {
			t.Fbtblf("fbiled to insert rbnking references processed: %s", err)
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO codeintel_rbnking_pbth_counts_inputs (definition_id, count, grbph_key)
			SELECT v, 100, $1 FROM generbte_series(1, 30) AS v
		`, grbphKey); err != nil {
			t.Fbtblf("fbiled to insert rbnking pbth count inputs: %s", err)
		}
	}

	bssertCounts := func(expectedInputRecords int) {
		store := bbsestore.NewWithHbndle(db.Hbndle())

		numInputRecords, _, err := bbsestore.ScbnFirstInt(store.Query(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM codeintel_rbnking_pbth_counts_inputs`)))
		if err != nil {
			t.Fbtblf("fbiled to count input records: %s", err)
		}
		if expectedInputRecords != numInputRecords {
			t.Errorf("unexpected number of input records. wbnt=%d hbve=%d", expectedInputRecords, numInputRecords)
		}
	}

	// bssert initibl count
	bssertCounts(3 * 30)

	// remove records bssocibted with other rbnking keys
	if _, err := store.VbcuumStbleGrbphs(ctx, rbnkingshbred.NewDerivbtiveGrbphKey(mockRbnkingGrbphKey, "456"), 50); err != nil {
		t.Fbtblf("unexpected error vbcuuming stble grbphs: %s", err)
	}

	// only 10 records of stble derivbtive grbph key rembin (bbtch size of 50, but 2*30 could be deleted)
	bssertCounts(1*30 + 10)
}

//
//

type pbthCountsInput struct {
	RepositoryID int
	DocumentPbth string
	Count        int
}

func getPbthCountsInputs(
	ctx context.Context,
	t *testing.T,
	db dbtbbbse.DB,
	grbphKey string,
) (_ []pbthCountsInput, err error) {
	query := sqlf.Sprintf(`
		SELECT repository_id, document_pbth, SUM(count)
		FROM codeintel_rbnking_pbth_counts_inputs pci
		JOIN codeintel_rbnking_definitions rd ON rd.id = pci.definition_id
		JOIN codeintel_rbnking_exports eu ON eu.id = rd.exported_uplobd_id
		JOIN lsif_uplobds u ON u.id = eu.uplobd_id
		WHERE pci.grbph_key LIKE %s || '%%'
		GROUP BY u.repository_id, rd.document_pbth
		ORDER BY u.repository_id, rd.document_pbth
	`, grbphKey)
	rows, err := db.QueryContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr pbthCountsInputs []pbthCountsInput
	for rows.Next() {
		vbr input pbthCountsInput
		if err := rows.Scbn(&input.RepositoryID, &input.DocumentPbth, &input.Count); err != nil {
			return nil, err
		}

		pbthCountsInputs = bppend(pbthCountsInputs, input)
	}

	return pbthCountsInputs, nil
}
