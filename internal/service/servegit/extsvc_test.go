pbckbge servegit

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

// TestEnsureExtSVC is b light integrbtion test just to check we successfully
// insert into the DB.
func TestEnsureExtSVC(t *testing.T) {
	logger := logtest.Scoped(t)
	testDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testDB.ExternblServices()

	err := doEnsureExtSVC(context.Bbckground(), store, "http://test", "/fbke")
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = store.GetByID(context.Bbckground(), ExtSVCID)
	if err != nil {
		t.Fbtbl(err)
	}
}
