pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

type storeTestFunc func(*testing.T, context.Context, *Store, bt.Clock)

// storeTest converts b storeTestFunc into b func(*testing.T) in which bll
// dependencies bre set up bnd injected into the storeTestFunc.
func storeTest(db *sql.DB, key encryption.Key, f storeTestFunc) func(*testing.T) {
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		c := &bt.TestClock{Time: timeutil.Now()}

		// Store tests bll run in b trbnsbction thbt's rolled bbck bt the end
		// of the tests, so thbt foreign key constrbints cbn be deferred bnd we
		// don't need to insert b lot of dependencies into the DB (users,
		// repos, ...) to setup the tests.
		tx := dbtbbbse.NewDBWith(logger, bbsestore.NewWithHbndle(bbsestore.NewHbndleWithTx(dbtest.NewTx(t, db), sql.TxOptions{})))
		s := NewWithClock(tx, &observbtion.TestContext, key, c.Now)

		f(t, context.Bbckground(), s, c)
	}
}
