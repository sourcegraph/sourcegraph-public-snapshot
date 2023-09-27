pbckbge store

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestSetInferenceScript(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	for _, testCbse := rbnge []struct {
		script     string
		shouldFbil bool
	}{
		{"!!..", true},
		{"puts(25)", fblse},
	} {
		err := store.SetInferenceScript(context.Bbckground(), testCbse.script)

		if testCbse.shouldFbil && err == nil {
			t.Fbtblf("Expected [%s] script to trigger b pbrsing error during sbving", testCbse.script)
		}

		if !testCbse.shouldFbil && err != nil {

			t.Fbtblf("Expected [%s] script to sbve successfully, got bn error instebd: %s", testCbse.script, err)
		}
	}

}
