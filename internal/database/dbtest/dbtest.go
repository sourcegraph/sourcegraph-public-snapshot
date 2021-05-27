package dbtest

import (
	crand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
)

// NewTx opens a transaction off of the given database, returning that
// transaction if an error didn't occur.
//
// After opening this transaction, it executes the query
//     SET CONSTRAINTS ALL DEFERRED
// which aids in testing.
func NewTx(t testing.TB, db *sql.DB) *sql.Tx {
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback()
	})

	return tx
}

// recycleC is a channel that is used to recycle migrated databases
var recycleC = make(chan *sql.DB)
var inflightDBs int32

func tryGetRecycledDB(t testing.TB) (*sql.DB, bool) {
	// Only attempt to re-use a migrated database if there are more than 4 already created
	if atomic.LoadInt32(&inflightDBs) >= 4 {
		select {
		case migratedDB := <-recycleC:
			// If another test finishes with its database, truncate all tables and return it
			emptyDBPreserveSchema(t, migratedDB)
			return migratedDB, true
		case <-time.After(100 * time.Millisecond):
			// Only wait for 100 milliseconds before just creating a new database
		}
	}
	return nil, false
}

func tryPutRecycledDB(db *sql.DB) bool {
	select {
	case recycleC <- db:
		return true
	default:
		// The getter waits for 100 milliseconds,
		// so don't block cleanup trying to recycle a database
		return false
	}
}

// Use a shared, locked RNG to avoid issues with multiple concurrent tests getting
// the same random database number (unlikely, but has been observed).
// Use crypto/rand.Read() to use an OS source of entropy, since, against all odds,
// using nanotime was causing conflicts.
var rng = rand.New(rand.NewSource(func() int64 {
	b := [8]byte{}
	if _, err := crand.Read(b[:]); err != nil {
		panic(err)
	}
	return int64(binary.LittleEndian.Uint64(b[:]))
}()))
var rngLock sync.Mutex

// NewDB returns a connection to a clean, new temporary testing database
// with the same schema as Sourcegraph's production Postgres database.
func NewDB(t testing.TB, dsn string) *sql.DB {
	if recycled, ok := tryGetRecycledDB(t); ok {
		return recycled
	}

	var err error
	var config *url.URL
	if dsn == "" {
		config, err = url.Parse("postgres://127.0.0.1/?sslmode=disable&timezone=UTC")
		if err != nil {
			t.Fatalf("failed to parse dsn %q: %s", dsn, err)
		}
		updateDSNFromEnv(config)
	} else {
		config, err = url.Parse(dsn)
		if err != nil {
			t.Fatalf("failed to parse dsn %q: %s", dsn, err)
		}
	}

	initTemplateDB(t, config)

	rngLock.Lock()
	dbname := "sourcegraph-test-" + strconv.FormatUint(rng.Uint64(), 10)
	rngLock.Unlock()

	db := dbConn(t, config)
	dbExec(t, db, `CREATE DATABASE `+pq.QuoteIdentifier(dbname)+` TEMPLATE `+pq.QuoteIdentifier(templateDBName()))

	config.Path = "/" + dbname
	testDB := dbConn(t, config)
	testDB.SetMaxOpenConns(3)

	t.Cleanup(func() {
		defer db.Close()

		if t.Failed() {
			t.Logf("DATABASE %s left intact for inspection", dbname)
			return
		}

		if tryPutRecycledDB(testDB) {
			return
		}

		if err := testDB.Close(); err != nil {
			t.Fatalf("failed to close test database: %s", err)
		}
		atomic.AddInt32(&inflightDBs, -1)
		dbExec(t, db, killClientConnsQuery, dbname)
		dbExec(t, db, `DROP DATABASE `+pq.QuoteIdentifier(dbname))
	})

	atomic.AddInt32(&inflightDBs, 1)
	return testDB
}

var templateOnce sync.Once

// initTemplateDB creates a template database with a fully migrated schema for the
// current package. New databases can then do a cheap copy of the migrated schema
// rather than running the full migration every time.
func initTemplateDB(t testing.TB, config *url.URL) {
	templateOnce.Do(func() {
		templateName := templateDBName()
		db := dbConn(t, config)
		_, err := db.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(templateName))
		if err != nil {
			pgErr := &pgconn.PgError{}
			if errors.As(err, &pgErr) && pgErr.Code == "42P04" {
				// Ignore database already exists errors.
				// Postgres doesn't support CREATE DATABASE IF NOT EXISTS,
				// so we just try to create it, and ignore the error if it's
				// because the database already exists.
			} else {
				t.Fatalf("Failed to create database: %s", err)
			}
		}
		defer db.Close()

		cfgCopy := *config
		cfgCopy.Path = "/" + templateName
		templateDB := dbConn(t, &cfgCopy)
		defer templateDB.Close()

		for _, database := range []*dbconn.Database{
			dbconn.Frontend,
			dbconn.CodeIntel,
		} {
			m, err := dbconn.NewMigrate(templateDB, database)
			if err != nil {
				t.Fatalf("failed to construct migrations: %s", err)
			}
			defer m.Close()
			if err = dbconn.DoMigrate(m); err != nil {
				t.Fatalf("failed to apply migrations: %s", err)
			}
		}
	})
}

// templateDBName returns the name of the template database
// for the currently running package.
func templateDBName() string {
	return "sourcegraph-test-template-" + wdHash()
}

// wdHash returns a hash of the current working directory.
// This is useful to get a stable identifier for the package running
// the tests.
func wdHash() string {
	h := fnv.New64()
	wd, _ := os.Getwd()
	h.Write([]byte(wd))
	return strconv.Itoa(int(h.Sum64()))
}

func dbConn(t testing.TB, cfg *url.URL) *sql.DB {
	db, err := dbconn.NewRaw(cfg.String())
	if err != nil {
		t.Fatalf("failed to connect to database %q: %s", cfg, err)
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

// updateDSNFromEnv updates dsn based on PGXXX environment variables set on
// the frontend.
func updateDSNFromEnv(dsn *url.URL) {
	if host := os.Getenv("PGHOST"); host != "" {
		dsn.Host = host
	}

	if port := os.Getenv("PGPORT"); port != "" {
		dsn.Host += ":" + port
	}

	if user := os.Getenv("PGUSER"); user != "" {
		if password := os.Getenv("PGPASSWORD"); password != "" {
			dsn.User = url.UserPassword(user, password)
		} else {
			dsn.User = url.User(user)
		}
	}

	if db := os.Getenv("PGDATABASE"); db != "" {
		dsn.Path = db
	}

	if sslmode := os.Getenv("PGSSLMODE"); sslmode != "" {
		qry := dsn.Query()
		qry.Set("sslmode", sslmode)
		dsn.RawQuery = qry.Encode()
	}
}

func emptyDBPreserveSchema(t testing.TB, d *sql.DB) {
	_, err := d.Exec(`SELECT * FROM schema_migrations`)
	if err != nil {
		t.Fatalf("Table schema_migrations not found: %v", err)
	}

	var conds []string
	conds = append(conds, fmt.Sprintf("table_name != '%s'", dbconn.Frontend.MigrationsTable))
	conds = append(conds, fmt.Sprintf("table_name != '%s'", dbconn.CodeIntel.MigrationsTable))

	rows, err := d.Query("SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' AND " + strings.Join(conds, " AND "))
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, table)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if testing.Verbose() {
		t.Logf("Truncating all %d tables", len(tables))
	}
	_, err = d.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY")
	if err != nil {
		t.Fatal(err)
	}
}
