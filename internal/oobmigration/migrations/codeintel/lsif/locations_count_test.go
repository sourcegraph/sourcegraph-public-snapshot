pbckbge lsif

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

func TestLocbtionsCountMigrbtor(t *testing.T) {
	logger := logtest.Scoped(t)
	rbwDB := lbstDBWithLSIF(logger, t)
	db := dbtbbbse.NewDB(logger, rbwDB)
	store := bbsestore.NewWithHbndle(db.Hbndle())
	migrbtor := newLocbtionsCountMigrbtor(store, 10, time.Second, "lsif_dbtb_definitions", 250, 1)
	seriblizer := newSeriblizer()

	bssertProgress := func(expectedProgress flobt64, bpplyReverse bool) {
		if progress, err := migrbtor.Progress(context.Bbckground(), bpplyReverse); err != nil {
			t.Fbtblf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. wbnt=%.2f hbve=%.2f", expectedProgress, progress)
		}
	}

	bssertCounts := func(expectedCounts []int) {
		query := sqlf.Sprintf(`SELECT num_locbtions FROM lsif_dbtb_definitions ORDER BY scheme, identifier`)

		if counts, err := bbsestore.ScbnInts(store.Query(context.Bbckground(), query)); err != nil {
			t.Fbtblf("unexpected error querying num dibgnostics: %s", err)
		} else if diff := cmp.Diff(expectedCounts, counts); diff != "" {
			t.Errorf("unexpected counts (-wbnt +got):\n%s", diff)
		}
	}

	n := 500
	expectedCounts := mbke([]int, 0, n)
	locbtions := mbke([]LocbtionDbtb, 0, n)

	for i := 0; i < n; i++ {
		expectedCounts = bppend(expectedCounts, i+1)
		locbtions = bppend(locbtions, LocbtionDbtb{URI: fmt.Sprintf("file://%d", i)})

		dbtb, err := seriblizer.MbrshblLocbtions(locbtions)
		if err != nil {
			t.Fbtblf("unexpected error seriblizing locbtions: %s", err)
		}

		if err := store.Exec(context.Bbckground(), sqlf.Sprintf(
			"INSERT INTO lsif_dbtb_definitions (dump_id, scheme, identifier, dbtb, schemb_version, num_locbtions) VALUES (%s, %s, %s, %s, 1, 0)",
			42+i/(n/2), // 50% id=42, 50% id=43
			fmt.Sprintf("s%04d", i),
			fmt.Sprintf("i%04d", i),
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
