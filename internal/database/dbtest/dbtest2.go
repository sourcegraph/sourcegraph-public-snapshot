package dbtest

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
)

func NewFast(t testing.TB) *sql.DB {
	t.Helper()
	u, err := url.Parse(getDSN())
	if err != nil {
		t.Fatalf("failed to parse dsn: %s", err)
	}

	pool, err := newPoolFromURL(u)
	if err != nil {
		t.Fatalf("failed to create new pool: %s", err)
	}
	t.Cleanup(func() { pool.db.Close() })

	return newFromPool(t, u, pool)
}

const defaultDSN = `postgres://sourcegraph:sourcegraph@127.0.0.1:5432/postgres?sslmode=disable&timezone=UTC`

func getDSN() string {
	if dsn, ok := os.LookupEnv("PGDATASOURCE"); ok {
		return dsn
	}

	return defaultDSN
}

func NewFastWithDSN(t testing.TB, dsn string) *sql.DB {
	t.Helper()
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("failed to parse dsn: %s", err)
	}

	pool, err := newPoolFromURL(u)
	if err != nil {
		t.Fatalf("failed to create new pool: %s", err)
	}
	return newFromPool(t, u, pool)
}

func newFromPool(t testing.TB, u *url.URL, pool *testDatabasePool) *sql.DB {
	ctx := context.Background()
	tdb, err := pool.GetTemplate(ctx, u, dbconn.Frontend, dbconn.CodeIntel)
	if err != nil {
		t.Fatalf("failed to get or create template db: %s", err)
	}

	mdb, err := pool.GetMigratedDB(ctx, tdb)
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

	testDBURL := urlWithDB(u, mdb.Name)
	testDB := dbConn(t, testDBURL)
	t.Cleanup(func() { testDB.Close() })

	return testDB
}

func NewFastTx(t testing.TB) *sql.Tx {
	t.Helper()
	u, err := url.Parse(getDSN())
	if err != nil {
		t.Fatalf("failed to parse dsn: %s", err)
	}

	pool, err := newPoolFromURL(u)
	if err != nil {
		t.Fatalf("failed to create new pool: %s", err)
	}
	t.Cleanup(func() { pool.db.Close() })
}

func urlWithDB(u *url.URL, dbName string) *url.URL {
	uCopy := *u
	uCopy.Path = "/" + dbName
	return &uCopy
}

func newPoolFromURL(u *url.URL) (*testDatabasePool, error) {
	db, err := dbconn.NewRaw(u.String())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Ignore already exists error
	// TODO: return error if it's not an already exists error
	_, _ = db.Exec("CREATE DATABASE dbtest_pool")

	poolDBURL := urlWithDB(u, "dbtest_pool")
	poolDB, err := dbconn.NewRaw(poolDBURL.String())
	if err != nil {
		return nil, err
	}

	if !poolSchemaUpToDate(poolDB) {
		poolDB.Close()
		if _, err = db.Exec("DROP DATABASE dbtest_pool"); err != nil {
			return nil, err
		}
		if _, err = db.Exec("CREATE DATABASE dbtest_pool"); err != nil {
			return nil, err
		}

		poolDB, err = dbconn.NewRaw(poolDBURL.String())
		if err != nil {
			return nil, err
		}

		if err := migratePoolDB(poolDB); err != nil {
			return nil, err
		}
	}

	return &testDatabasePool{db: poolDB}, nil
}
