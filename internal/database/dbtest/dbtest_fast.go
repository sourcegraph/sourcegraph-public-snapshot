package dbtest

import (
	"context"
	"database/sql"
	"net/url"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/test"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewFastDB returns a clean database that will be deleted
// at the end of the test.
func NewFastDB(t testing.TB) *sql.DB {
	if testing.Short() {
		t.Skip("skipping DB test since -short specified")
	}
	t.Helper()

	pool, u, err := getDefaultPool()
	if err != nil {
		t.Fatalf("error getting pool: %s", err)
	}

	return newFromPool(t, u, pool)
}

// NewFastTx returns a transaction in a clean database. At the end of the test,
// the transaction will be rolled back, and the clean database can be reused
func NewFastTx(t testing.TB) *sql.Tx {
	if testing.Short() {
		t.Skip("skipping DB test since -short specified")
	}
	t.Helper()

	pool, u, err := getDefaultPool()
	if err != nil {
		t.Fatalf("error getting pool: %s", err)
	}

	return newTxFromPool(t, u, pool)
}

func Savepoint(t testing.TB, tx *sql.Tx) (rollback func()) {
	u, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("failed to create uuid: %s", err)
	}

	_, err = tx.Exec("SAVEPOINT " + pq.QuoteIdentifier(u.String()))
	if err != nil {
		t.Fatalf("failed to create savepoint: %s", err)
	}
	return func() {
		_, err := tx.Exec("ROLLBACK TO SAVEPOINT " + pq.QuoteIdentifier(u.String()))
		if err != nil {
			t.Fatalf("failed to roll back: %s", err)
		}
	}
}

var (
	defaultOnce sync.Once
	defaultPool *testDatabasePool
	defaultURL  *url.URL
	defaultErr  error
)

// getDefaultPool returns a pool initialized once per process. This keeps
// us from opening a ton of parallel database connections per process.
func getDefaultPool() (*testDatabasePool, *url.URL, error) {
	defaultOnce.Do(func() {
		defaultURL, defaultErr = getDSN()
		if defaultErr != nil {
			return
		}

		defaultPool, defaultErr = newPoolFromURL(defaultURL)
		if defaultErr != nil {
			return
		}

		defaultErr = defaultPool.CleanUpOldDBs(context.Background(), schemas.Frontend, schemas.CodeIntel)
	})
	return defaultPool, defaultURL, defaultErr
}

func newFromPool(t testing.TB, u *url.URL, pool *testDatabasePool) *sql.DB {
	ctx := context.Background()

	// Get or create the template database
	tdb, err := pool.GetTemplate(ctx, u, schemas.Frontend, schemas.CodeIntel)
	if err != nil {
		t.Fatalf("failed to get or create template db: %s", err)
	}

	// Get or create a database cloned from the template database
	mdb, err := pool.GetMigratedDB(ctx, false, tdb)
	if err != nil {
		t.Fatalf("failed to get or create migrated db: %s", err)
	}
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("database %q left intact for inspection", mdb.Name)
			return
		}

		err := pool.DeleteMigratedDB(ctx, mdb)
		if err != nil {
			t.Fatalf("failed to clean up migrated db %q: %s", mdb.Name, err)
		}
	})

	// Open a connection to the clean database
	testDBURL := urlWithDB(u, mdb.Name)
	testDB := dbConn(t, testDBURL)
	// Some tests that exercise concurrency need lots of connections or they block forever.
	// e.g. TestIntegration/DBStore/Syncer/MultipleServices
	testDB.SetMaxOpenConns(10)
	t.Cleanup(func() { testDB.Close() })

	return testDB
}

func newTxFromPool(t testing.TB, u *url.URL, pool *testDatabasePool) *sql.Tx {
	ctx := context.Background()
	tdb, err := pool.GetTemplate(ctx, u, schemas.Frontend, schemas.CodeIntel)
	if err != nil {
		t.Fatalf("failed to get or create template db: %s", err)
	}

	mdb, err := pool.GetMigratedDB(ctx, true, tdb)
	if err != nil {
		t.Fatalf("failed to get or create migrated db: %s", err)
	}
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("database %q left intact for inspection", mdb.Name)
			return
		}

		err := pool.PutMigratedDB(ctx, mdb)
		if err != nil {
			t.Fatalf("failed to unclaim migrated db %q: %s", mdb.Name, err)
		}
	})

	testDBURL := urlWithDB(u, mdb.Name)
	testDB := dbConn(t, testDBURL)
	t.Cleanup(func() { testDB.Close() })

	tx, err := testDB.Begin()
	if err != nil {
		t.Fatalf("failed to create a transaction: %s", err)
	}
	t.Cleanup(func() { tx.Rollback() })
	return tx
}

func urlWithDB(u *url.URL, dbName string) *url.URL {
	uCopy := *u
	uCopy.Path = "/" + dbName
	return &uCopy
}

func newPoolFromURL(u *url.URL) (_ *testDatabasePool, err error) {
	db, err := connections.NewTestDB(u.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	// Ignore already exists error
	// TODO: return error if it's not an already exists error
	_, _ = db.Exec("CREATE DATABASE dbtest_pool")

	poolDBURL := urlWithDB(u, "dbtest_pool")
	poolDB, err := connections.NewTestDB(poolDBURL.String())
	if err != nil {
		return nil, err
	}

	if !poolSchemaUpToDate(poolDB) {
		if err := poolDB.Close(); err != nil {
			return nil, err
		}

		if _, err = db.Exec("DROP DATABASE dbtest_pool"); err != nil {
			return nil, err
		}
		if _, err = db.Exec("CREATE DATABASE dbtest_pool"); err != nil {
			return nil, err
		}

		poolDB, err = connections.NewTestDB(poolDBURL.String())
		if err != nil {
			return nil, err
		}

		if err := migratePoolDB(poolDB); err != nil {
			return nil, err
		}
	}

	return newTestDatabasePool(poolDB), nil
}
