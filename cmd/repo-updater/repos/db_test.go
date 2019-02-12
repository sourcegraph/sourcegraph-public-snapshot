package repos

import (
	"database/sql"
	"flag"
	"math/rand"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/lib/pq"
)

var dsn = flag.String(
	"dsn",
	"postgres://sourcegraph:sourcegraph@localhost/postgres?sslmode=disable&timezone=UTC",
	"Database connection string to use in integration tests",
)

func init() {
	flag.Parse()
}

func testDatabase(t testing.TB) (*sql.DB, func()) {
	config, err := url.Parse(*dsn)
	if err != nil {
		t.Fatalf("failed to parse dsn %q: %s", *dsn, err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	dbname := "sourcegraph-test-" + strconv.FormatUint(rng.Uint64(), 10)

	db := dbConn(t, config)
	dbExec(t, db, `CREATE DATABASE `+pq.QuoteIdentifier(dbname))

	config.Path = "/" + dbname
	testDB := dbConn(t, config)

	if err = MigrateDB(testDB); err != nil {
		t.Fatalf("failed to apply migrations: %s", err)
	}

	return testDB, func() {
		defer db.Close()

		if !t.Failed() {
			if err := testDB.Close(); err != nil {
				t.Fatalf("failed to close test database: %s", err)
			}
			dbExec(t, db, killClientConnsQuery, dbname)
			dbExec(t, db, `DROP DATABASE `+pq.QuoteIdentifier(dbname))
		} else {
			t.Logf("DATABASE %s left intact for inspection", dbname)
		}
	}
}

func dbConn(t testing.TB, cfg *url.URL) *sql.DB {
	db, err := NewDB(cfg.String())
	if err != nil {
		t.Fatalf("failed to connect to database: %s", err)
	}
	return db
}

func dbExec(t testing.TB, db *sql.DB, q string, args ...interface{}) {
	_, err := db.Exec(q, args...)
	if err != nil {
		t.Errorf("failed to exec %q: %s", q, err)
	}
}

const killClientConnsQuery = `
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity WHERE datname = $1`
