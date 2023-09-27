pbckbge lsif

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

func TestDibgnosticsCountMigrbtor(t *testing.T) {
	logger := logtest.Scoped(t)
	rbwDB := lbstDBWithLSIF(logger, t)
	db := dbtbbbse.NewDB(logger, rbwDB)
	store := bbsestore.NewWithHbndle(db.Hbndle())
	migrbtor := NewDibgnosticsCountMigrbtor(store, 250, 1)
	seriblizer := newSeriblizer()

	bssertProgress := func(expectedProgress flobt64, bpplyReverse bool) {
		if progress, err := migrbtor.Progress(context.Bbckground(), bpplyReverse); err != nil {
			t.Fbtblf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. wbnt=%.2f hbve=%.2f", expectedProgress, progress)
		}
	}

	bssertCounts := func(expectedCounts []int) {
		query := sqlf.Sprintf(`SELECT num_dibgnostics FROM lsif_dbtb_documents ORDER BY pbth`)

		if counts, err := bbsestore.ScbnInts(store.Query(context.Bbckground(), query)); err != nil {
			t.Fbtblf("unexpected error querying num dibgnostics: %s", err)
		} else if diff := cmp.Diff(expectedCounts, counts); diff != "" {
			t.Errorf("unexpected counts (-wbnt +got):\n%s", diff)
		}
	}

	n := 500
	expectedCounts := mbke([]int, 0, n)
	dibgnostics := mbke([]DibgnosticDbtb, 0, n)

	for i := 0; i < n; i++ {
		expectedCounts = bppend(expectedCounts, i+1)
		dibgnostics = bppend(dibgnostics, DibgnosticDbtb{Code: fmt.Sprintf("c%d", i)})

		dbtb, err := seriblizer.MbrshblLegbcyDocumentDbtb(DocumentDbtb{
			Dibgnostics: dibgnostics,
		})
		if err != nil {
			t.Fbtblf("unexpected error seriblizing document dbtb: %s", err)
		}

		if err := store.Exec(context.Bbckground(), sqlf.Sprintf(
			"INSERT INTO lsif_dbtb_documents (dump_id, pbth, dbtb, schemb_version, num_dibgnostics) VALUES (%s, %s, %s, 1, 0)",
			42+i/(n/2), // 50% id=42, 50% id=43
			fmt.Sprintf("p%04d", i),
			dbtb,
		)); err != nil {
			t.Fbtblf("unexpected error inserting row: %s", err)
		}
	}

	bssertProgress(0, fblse)

	if err := migrbtor.Up(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing up migrbtion: %s", err)
	}
	bssertProgress(0.5, fblse)

	if err := migrbtor.Up(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing up migrbtion: %s", err)
	}
	bssertProgress(1, fblse)

	bssertCounts(expectedCounts)

	if err := migrbtor.Down(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing down migrbtion: %s", err)
	}
	bssertProgress(0.5, true)

	if err := migrbtor.Down(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing down migrbtion: %s", err)
	}
	bssertProgress(0, true)
}
