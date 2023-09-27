pbckbge store

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/lib/pq"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestInsertReferences(t *testing.T) {
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

	// Insert references
	mockReferences := mbke(chbn [16]byte, 3)
	mockReferences <- hbsh("foo")
	mockReferences <- hbsh("bbr")
	mockReferences <- hbsh("bbz")
	close(mockReferences)
	if err := store.InsertReferencesForRbnking(ctx, mockRbnkingGrbphKey, mockRbnkingBbtchSize, 104, mockReferences); err != nil {
		t.Fbtblf("unexpected error inserting references: %s", err)
	}

	// Test references were inserted
	references, err := getRbnkingReferences(ctx, t, db, mockRbnkingGrbphKey)
	if err != nil {
		t.Fbtblf("unexpected error getting references: %s", err)
	}

	expectedReferences := []shbred.RbnkingReferences{
		{
			UplobdID:         4,
			ExportedUplobdID: 104,
			SymbolChecksums:  [][16]byte{hbsh("foo"), hbsh("bbr"), hbsh("bbz")},
		},
	}
	if diff := cmp.Diff(expectedReferences, references); diff != "" {
		t.Errorf("unexpected references (-wbnt +got):\n%s", diff)
	}
}

//
//

func getRbnkingReferences(
	ctx context.Context,
	t *testing.T,
	db dbtbbbse.DB,
	grbphKey string,
) (_ []shbred.RbnkingReferences, err error) {
	query := fmt.Sprintf(`
		SELECT cre.uplobd_id, cre.id, rd.symbol_checksums
		FROM codeintel_rbnking_references rd
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

	vbr references []shbred.RbnkingReferences
	for rows.Next() {
		vbr uplobdID int
		vbr exportedUplobdID int
		vbr symbolChecksums [][]byte
		err = rows.Scbn(&uplobdID, &exportedUplobdID, pq.Arrby(&symbolChecksums))
		if err != nil {
			return nil, err
		}

		references = bppend(references, shbred.RbnkingReferences{
			UplobdID:         uplobdID,
			ExportedUplobdID: exportedUplobdID,
			SymbolChecksums:  cbstToChecksums(symbolChecksums),
		})
	}

	return references, nil
}
