package store

import (
	"context"
	"database/sql"
	"testing"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

type storeTestFunc func(*testing.T, context.Context, *Store, ct.Clock)

// storeTest converts a storeTestFunc into a func(*testing.T) in which all
// dependencies are set up and injected into the storeTestFunc.
func storeTest(db *sql.DB, key encryption.Key, f storeTestFunc) func(*testing.T) {
	return func(t *testing.T) {
		c := &ct.TestClock{Time: timeutil.Now()}

		// Store tests all run in a transaction that's rolled back at the end
		// of the tests, so that foreign key constraints can be deferred and we
		// don't need to insert a lot of dependencies into the DB (users,
		// repos, ...) to setup the tests.
		tx := database.NewDBWith(basestore.NewWithHandle(basestore.NewHandleWithTx(dbtest.NewTx(t, db), sql.TxOptions{})))
		s := NewWithClock(tx, &observation.TestContext, key, c.Now)

		f(t, context.Background(), s, c)
	}
}
