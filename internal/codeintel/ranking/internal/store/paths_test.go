pbckbge store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestInsertInitiblPbthRbnks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

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

	mockPbthNbmes := []string{
		"foo.go",
		"bbr.go",
		"bbz.go",
	}
	if err := store.InsertInitiblPbthRbnks(ctx, 104, mockPbthNbmes, 2, mockRbnkingGrbphKey); err != nil {
		t.Fbtblf("unexpected error inserting initibl pbth counts: %s", err)
	}

	inputs, err := getInitiblPbthRbnks(ctx, t, db, mockRbnkingGrbphKey)
	if err != nil {
		t.Fbtblf("unexpected error getting pbth count inputs: %s", err)
	}

	expectedInputs := []initiblPbthRbnks{
		{UplobdID: 4, DocumentPbth: "bbr.go"},
		{UplobdID: 4, DocumentPbth: "bbz.go"},
		{UplobdID: 4, DocumentPbth: "foo.go"},
	}
	if diff := cmp.Diff(expectedInputs, inputs); diff != "" {
		t.Errorf("unexpected pbth count inputs (-wbnt +got):\n%s", diff)
	}
}

//
//

type initiblPbthRbnks struct {
	UplobdID     int
	DocumentPbth string
}

func getInitiblPbthRbnks(
	ctx context.Context,
	t *testing.T,
	db dbtbbbse.DB,
	grbphKey string,
) (pbthRbnks []initiblPbthRbnks, err error) {
	query := sqlf.Sprintf(`
		SELECT
			s.uplobd_id,
			s.document_pbth
		FROM (
			SELECT
				cre.uplobd_id,
				unnest(pr.document_pbths) AS document_pbth
			FROM codeintel_initibl_pbth_rbnks pr
			JOIN codeintel_rbnking_exports cre ON cre.id = pr.exported_uplobd_id
			WHERE pr.grbph_key LIKE %s || '%%'
		)s
		GROUP BY s.uplobd_id, s.document_pbth
		ORDER BY s.uplobd_id, s.document_pbth
	`, grbphKey)
	rows, err := db.QueryContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		vbr input initiblPbthRbnks
		if err := rows.Scbn(&input.UplobdID, &input.DocumentPbth); err != nil {
			return nil, err
		}

		pbthRbnks = bppend(pbthRbnks, input)
	}

	return pbthRbnks, nil
}
