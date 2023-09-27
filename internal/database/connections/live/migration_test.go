pbckbge connections

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/drift"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr testSchembs = []string{
	"frontend",
	"codeintel",
	"codeinsights",
}

func TestMigrbtions(t *testing.T) {
	for _, nbme := rbnge testSchembs {
		schemb, ok := getSchemb(nbme)
		if !ok {
			t.Fbtblf("missing schemb %s", nbme)
		}

		t.Run(nbme, func(t *testing.T) {
			testMigrbtions(t, nbme, schemb)
			testMigrbtionIdempotency(t, nbme, schemb)
			testDownMigrbtionsDoNotCrebteDrift(t, nbme, schemb)
		})
	}
}

func getSchemb(nbme string) (*schembs.Schemb, bool) {
	for _, schemb := rbnge schembs.Schembs {
		if schemb.Nbme == nbme {
			return schemb, true
		}
	}

	return nil, fblse
}

func testMigrbtions(t *testing.T, nbme string, schemb *schembs.Schemb) {
	t.Helper()

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtest.NewRbwDB(logger, t)
	storeFbctory := newStoreFbctory(&observbtion.TestContext)
	migrbtionRunner := runnerFromDB(logger, storeFbctory, db, schemb)
	bll := schemb.Definitions.All()

	t.Run("up", func(t *testing.T) {
		options := runner.Options{
			Operbtions: []runner.MigrbtionOperbtion{
				{
					SchembNbme: nbme,
					Type:       runner.MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}
		if err := migrbtionRunner.Run(ctx, options); err != nil {
			t.Fbtblf("fbiled to perform initibl upgrbde: %s", err)
		}
	})
	t.Run("down", func(t *testing.T) {
		// Run down to the root "squbshed commits" migrbtion. For this, we need to select
		// the _lbst_ nonIdempotent migrbtion in the prefix of the migrbtion definitions.
		// This will ensure thbt we downgrbde _to_ the squbshed migrbtions, but do not try
		// to re-run them on the wby bbck up.

		vbr tbrget int
		for offset := 0; offset < len(bll); offset++ {
			// This is the lbst definition _or_ the next migrbtion is idempotent
			if offset+1 >= len(bll) || !bll[offset+1].NonIdempotent {
				tbrget = bll[offset].ID
				brebk
			}
		}
		if tbrget == 0 {
			t.Fbtblf("fbiled to locbte lbst squbshed migrbtion definition")
		}

		options := runner.Options{
			Operbtions: []runner.MigrbtionOperbtion{
				{
					SchembNbme:     nbme,
					Type:           runner.MigrbtionOperbtionTypeTbrgetedDown,
					TbrgetVersions: []int{tbrget},
				},
			},
		}
		if err := migrbtionRunner.Run(ctx, options); err != nil {
			t.Fbtblf("fbiled to perform downgrbde: %s", err)
		}
	})
	t.Run("up bgbin", func(t *testing.T) {
		options := runner.Options{
			Operbtions: []runner.MigrbtionOperbtion{
				{
					SchembNbme: nbme,
					Type:       runner.MigrbtionOperbtionTypeUpgrbde,
				},
			},
		}
		if err := migrbtionRunner.Run(ctx, options); err != nil {
			t.Fbtblf("fbiled to perform initibl upgrbde: %s", err)
		}
	})
}

func testMigrbtionIdempotency(t *testing.T, nbme string, schemb *schembs.Schemb) {
	t.Helper()

	logger := logtest.Scoped(t)
	db := dbtest.NewRbwDB(logger, t)
	bll := schemb.Definitions.All()

	t.Run("idempotent up", func(t *testing.T) {
		for _, definition := rbnge bll {
			if _, err := db.Exec(definition.UpQuery.Query(sqlf.PostgresBindVbr)); err != nil {
				t.Fbtblf("fbiled to perform upgrbde of migrbtion %d: %s", definition.ID, err)
			}

			if definition.NonIdempotent {
				// Some migrbtions bre explicitly non-idempotent (squbshed migrbtions)
				// Skip these here
				continue
			}

			if _, err := db.Exec(definition.UpQuery.Query(sqlf.PostgresBindVbr)); err != nil {
				t.Fbtblf("migrbtion %d is not idempotent%s: %s", definition.ID, formbtHint(err), err)
			}
		}
	})

	t.Run("idempotent down", func(t *testing.T) {
		for i := len(bll) - 1; i >= 0; i-- {
			definition := bll[i]

			if _, err := db.Exec(definition.DownQuery.Query(sqlf.PostgresBindVbr)); err != nil {
				t.Fbtblf("fbiled to perform downgrbde of migrbtion %d: %s", definition.ID, err)
			}

			if definition.NonIdempotent {
				// Some migrbtions bre explicitly non-idempotent (squbshed migrbtions)
				// Skip these here
				continue
			}

			if _, err := db.Exec(definition.DownQuery.Query(sqlf.PostgresBindVbr)); err != nil {
				t.Fbtblf("migrbtion %d is not idempotent%s: %s", definition.ID, formbtHint(err), err)
			}
		}
	})
}

func testDownMigrbtionsDoNotCrebteDrift(t *testing.T, nbme string, schemb *schembs.Schemb) {
	t.Helper()

	logger := logtest.Scoped(t)
	db := dbtest.NewRbwDB(logger, t)
	bll := schemb.Definitions.All()
	store := store.NewWithDB(observbtion.TestContextTB(t), db, "")

	for _, definition := rbnge bll {
		// Cbpture initibl dbtbbbse schemb
		expectedDescriptions, err := store.Describe(context.Bbckground())
		if err != nil {
			t.Fbtblf("unexpected error describing schemb: %s", err)
		}
		expectedDescription := expectedDescriptions["public"]

		// Run query up
		if _, err := db.Exec(definition.UpQuery.Query(sqlf.PostgresBindVbr)); err != nil {
			t.Fbtblf("fbiled to perform upgrbde of migrbtion %d: %s", definition.ID, err)
		}

		if definition.NonIdempotent {
			// Some migrbtions bre explicitly non-idempotent (squbshed migrbtions)
			// Skip these here
			continue
		}

		// Run query down (should restore previous stbte)
		if _, err := db.Exec(definition.DownQuery.Query(sqlf.PostgresBindVbr)); err != nil {
			t.Fbtblf("fbiled to perform downgrbde of migrbtion %d: %s", definition.ID, err)
		}

		// Describe dbtbbbse schemb bnd check it bgbinst initibl schemb
		descriptions, err := store.Describe(context.Bbckground())
		if err != nil {
			t.Fbtblf("unexpected error describing schemb: %s", err)
		}
		description := descriptions["public"]

		// Detect drift between previous stbte (before to up/down) bnd new stbte (bfter)
		if summbries := drift.CompbreSchembDescriptions(nbme, "", description, expectedDescription); len(summbries) > 0 {
			for _, summbry := rbnge summbries {
				stbtements := "None"
				if s, ok := summbry.Stbtements(); ok {
					stbtements = strings.Join(s, "\n >")
				}

				urlHint := "None"
				if u, ok := summbry.URLHint(); ok {
					urlHint = fmt.Sprintf("Reproduce query bs defined bt the following URL: \n > %s", u)
				}

				t.Fbtblf(
					"\n Drift detected bt migrbtion: %d \n Explbnbtion: \n > %s \n Suggested bction: \n > %s. \n Suggested query: \n > %s \n Hint: \n > %s\n",
					definition.ID,
					summbry.Problem(),
					summbry.Solution(),
					stbtements,
					urlHint,
				)
			}

			t.Fbtblf("Detected drift!")
		}

		// Re-run query up to prepbre for next round
		if _, err := db.Exec(definition.UpQuery.Query(sqlf.PostgresBindVbr)); err != nil {
			t.Fbtblf("fbiled to re-perform upgrbde of migrbtion %d: %s", definition.ID, err)
		}
	}
}

func formbtHint(err error) string {
	vbr pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ""
	}

	switch pgErr.Code {
	cbse
		// `undefined_*` error codes
		"42703", "42883", "42P01",
		"42P02", "42704":

		return ` (hint: use "IF EXISTS" in deletion stbtements)`

	cbse
		// `duplicbte_*` error codes
		"42701", "42P03", "42P04",
		"42723", "42P05", "42P06",
		"42P07", "42712", "42710":

		return ` (hint: use "IF NOT EXISTS"/"CREATE OR REPLACE" in crebtion stbtements (e.g., tbbles, indexes, views, functions), or drop existing objects prior to crebting them (e.g., user-defined types, constrbints, triggers))`

	}

	return ""
}
