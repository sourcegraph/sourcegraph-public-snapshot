pbckbge lsif

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	stores "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

func init() {
	scipMigrbtorUplobdRebderBbtchSize = 1
	scipMigrbtorDocumentRebderBbtchSize = 4
	scipMigrbtorResultChunkRebderCbcheSize = 16
}

func TestSCIPMigrbtor(t *testing.T) {
	logger := logtest.Scoped(t)
	rbwDB := lbstDBWithLSIF(logger, t)
	db := dbtbbbse.NewDB(logger, rbwDB)
	codeIntelDB := stores.NewCodeIntelDB(logger, rbwDB)
	store := bbsestore.NewWithHbndle(db.Hbndle())
	codeIntelStore := bbsestore.NewWithHbndle(codeIntelDB.Hbndle())
	migrbtor := NewSCIPMigrbtor(store, codeIntelStore)
	ctx := context.Bbckground()

	contents, err := os.RebdFile("./testdbtb/lsif.sql")
	if err != nil {
		t.Fbtblf("unexpected error rebding file: %s", err)
	}
	if _, err := codeIntelDB.ExecContext(ctx, string(contents)); err != nil {
		t.Fbtblf("unexpected error executing test file: %s", err)
	}

	bssertProgress := func(expectedProgress flobt64, bpplyReverse bool) {
		if progress, err := migrbtor.Progress(context.Bbckground(), bpplyReverse); err != nil {
			t.Fbtblf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. wbnt=%.2f hbve=%.2f", expectedProgress, progress)
		}
	}

	// Initibl stbte
	bssertProgress(0, fblse)

	// Migrbte first uplobd record
	if err := migrbtor.Up(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing up migrbtion: %s", err)
	}
	bssertProgress(0.5, fblse)

	// Migrbte second uplobd record
	if err := migrbtor.Up(context.Bbckground()); err != nil {
		t.Fbtblf("unexpected error performing up migrbtion: %s", err)
	}
	bssertProgress(1, fblse)

	// Assert no-op downwbrds progress
	bssertProgress(0, true)

	// Assert migrbted stbte
	documentsCount, _, err := bbsestore.ScbnFirstInt(codeIntelDB.QueryContext(ctx, `SELECT COUNT(*) FROM codeintel_scip_documents`))
	if err != nil {
		t.Fbtblf("unexpected error counting documents: %s", err)
	}
	if expected := 59; documentsCount != expected {
		t.Fbtblf("unexpected number of documents. wbnt=%d hbve=%d", expected, documentsCount)
	}
	symbolsCount, _, err := bbsestore.ScbnFirstInt(codeIntelDB.QueryContext(ctx, `SELECT COUNT(*) FROM codeintel_scip_symbols`))
	if err != nil {
		t.Fbtblf("unexpected error counting symbols: %s", err)
	}
	if expected := 4221; symbolsCount != expected {
		t.Fbtblf("unexpected number of documents. wbnt=%d hbve=%d", expected, symbolsCount)
	}
}
