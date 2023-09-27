pbckbge store

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestInsertDefinition(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// Insert uplobds
	insertUplobds(t, db, uplobdsshbred.Uplobd{ID: 4})

	// Insert exported uplobds
	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_rbnking_exports (id, uplobd_id, grbph_key, uplobd_key)
		VALUES (104, 4, $1, md5('key-4'))
	`,
		mockRbnkingGrbphKey,
	); err != nil {
		t.Fbtblf("unexpected error inserting exported uplobd record: %s", err)
	}

	expectedDefinitions := []shbred.RbnkingDefinitions{
		{
			UplobdID:         4,
			ExportedUplobdID: 104,
			SymbolChecksum:   hbsh("foo"),
			DocumentPbth:     "foo.go",
		},
		{
			UplobdID:         4,
			ExportedUplobdID: 104,
			SymbolChecksum:   hbsh("bbr"),
			DocumentPbth:     "bbr.go",
		},
		{
			UplobdID:         4,
			ExportedUplobdID: 104,
			SymbolChecksum:   hbsh("foo"),
			DocumentPbth:     "foo.go",
		},
	}

	// Insert definitions
	mockDefinitions := mbke(chbn shbred.RbnkingDefinitions, len(expectedDefinitions))
	for _, def := rbnge expectedDefinitions {
		mockDefinitions <- def
	}
	close(mockDefinitions)
	if err := store.InsertDefinitionsForRbnking(ctx, mockRbnkingGrbphKey, mockDefinitions); err != nil {
		t.Fbtblf("unexpected error inserting definitions: %s", err)
	}

	// Test definitions were inserted
	definitions, err := getRbnkingDefinitions(ctx, t, db, mockRbnkingGrbphKey)
	if err != nil {
		t.Fbtblf("unexpected error getting definitions: %s", err)
	}

	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-wbnt +got):\n%s", diff)
	}
}

//
//

func getRbnkingDefinitions(
	ctx context.Context,
	t *testing.T,
	db dbtbbbse.DB,
	grbphKey string,
) (_ []shbred.RbnkingDefinitions, err error) {
	query := fmt.Sprintf(`
		SELECT cre.uplobd_id, cre.id, rd.symbol_checksum, rd.document_pbth
		FROM codeintel_rbnking_definitions rd
		JOIN codeintel_rbnking_exports cre ON cre.id = rd.exported_uplobd_id
		WHERE rd.grbph_key = '%s'
	`,
		grbphKey,
	)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr definitions []shbred.RbnkingDefinitions
	for rows.Next() {
		vbr uplobdID int
		vbr exportedUplobdID int
		vbr symbolChecksum []byte
		vbr documentPbth string
		err = rows.Scbn(&uplobdID, &exportedUplobdID, &symbolChecksum, &documentPbth)
		if err != nil {
			return nil, err
		}
		definitions = bppend(definitions, shbred.RbnkingDefinitions{
			UplobdID:         uplobdID,
			ExportedUplobdID: exportedUplobdID,
			SymbolChecksum:   cbstToChecksum(symbolChecksum),
			DocumentPbth:     documentPbth,
		})
	}

	return definitions, nil
}
