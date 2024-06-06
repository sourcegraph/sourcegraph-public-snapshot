package dbtest

import (
	crand "crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/lib/pq"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/test"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
)

// NewTx opens a transaction off of the given database, returning that
// transaction if an error didn't occur.
//
// After opening this transaction, it executes the query
//
//	SET CONSTRAINTS ALL DEFERRED
//
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

// NewDB returns a connection to a clean, new temporary testing database with
// the same schema as Sourcegraph's production Postgres database.
func NewDB(t testing.TB) *sql.DB {
	logger := logtest.Scoped(t)
	return newDB(logger, t, "migrated", schemas.Frontend, schemas.CodeIntel)
}

// NewCodeintelDB returns a connection to a new clean temporary testing database
// with only the codeintel schema applied
func NewCodeintelDB(t testing.TB) *sql.DB {
	logger := logtest.Scoped(t)
	return newDB(logger, t, "migrated-codeintel", schemas.CodeIntel)
}

// NewDBAtRev returns a connection to a clean, new temporary testing database with
// the same schema as Sourcegraph's production Postgres database at the given revision.
func NewDBAtRev(logger log.Logger, t testing.TB, rev string) *sql.DB {
	return newDB(
		logger,
		t,
		fmt.Sprintf("migrated-%s", rev),
		getSchemaAtRev(t, "frontend", rev),
		getSchemaAtRev(t, "codeintel", rev),
	)
}

func getSchemaAtRev(t testing.TB, name, rev string) *schemas.Schema {
	schema, err := schemas.ResolveSchemaAtRev(name, rev)
	if err != nil {
		t.Fatalf("failed to resolve %q schema: %s", name, err)
	}

	return schema
}

// NewInsightsDB returns a connection to a clean, new temporary testing database with
// the same schema as Sourcegraph's CodeInsights production Postgres database.
func NewInsightsDB(logger log.Logger, t testing.TB) *sql.DB {
	return newDB(logger, t, "insights", schemas.CodeInsights)
}

// NewRawDB returns a connection to a clean, new temporary testing database.
func NewRawDB(logger log.Logger, t testing.TB) *sql.DB {
	return newDB(logger, t, "raw")
}

func newDB(logger log.Logger, t testing.TB, name string, schemas ...*schemas.Schema) *sql.DB {
	if testing.Short() {
		t.Skip("DB tests disabled since go test -short is specified")
	}

	onceByName(name).Do(func() { initTemplateDB(logger, t, name, schemas) })
	return newFromDSN(logger, t, name)
}

var (
	onceByNameMap   = map[string]*sync.Once{}
	onceByNameMutex sync.Mutex
)

func onceByName(name string) *sync.Once {
	onceByNameMutex.Lock()
	defer onceByNameMutex.Unlock()

	if once, ok := onceByNameMap[name]; ok {
		return once
	}

	once := new(sync.Once)
	onceByNameMap[name] = once
	return once
}

func newFromDSN(logger log.Logger, t testing.TB, templateNamespace string) *sql.DB {
	if testing.Short() {
		t.Skip("skipping DB test since -short specified")
	}

	config, err := GetDSN()
	if err != nil {
		t.Fatalf("failed to parse dsn: %s", err)
	}

	rngLock.Lock()
	dbname := "sourcegraph-test-" + strconv.FormatUint(rng.Uint64(), 10)
	rngLock.Unlock()

	db := dbConn(logger, t, config)
	dbExec(t, db, `CREATE DATABASE `+pq.QuoteIdentifier(dbname)+` TEMPLATE `+pq.QuoteIdentifier(templateDBName(templateNamespace)))

	config.Path = "/" + dbname
	testDB := dbConn(logger, t, config)
	t.Logf("testdb: %s", config.String())

	// Some tests that exercise concurrency need lots of connections or they block forever.
	// e.g. TestIntegration/DBStore/Syncer/MultipleServices
	conns, err := strconv.Atoi(os.Getenv("TESTDB_MAXOPENCONNS"))
	if err != nil || conns == 0 {
		conns = 20
	}
	testDB.SetMaxOpenConns(conns)
	testDB.SetMaxIdleConns(1) // Default is 2, and within tests, it's not that important to have more than one.

	t.Cleanup(func() {
		defer db.Close()

		if t.Failed() && os.Getenv("CI") != "true" {
			t.Logf("DATABASE %s left intact for inspection", dbname)
			return
		}

		if err := testDB.Close(); err != nil {
			t.Fatalf("failed to close test database: %s", err)
		}
		dbExec(t, db, killClientConnsQuery, dbname)
		dbExec(t, db, `DROP DATABASE `+pq.QuoteIdentifier(dbname))
	})

	return testDB
}

// initTemplateDB creates a template database with a fully migrated schema for the
// current package. New databases can then do a cheap copy of the migrated schema
// rather than running the full migration every time.
func initTemplateDB(logger log.Logger, t testing.TB, templateNamespace string, dbSchemas []*schemas.Schema) {
	config, err := GetDSN()
	if err != nil {
		t.Fatalf("failed to parse dsn: %s", err)
	}

	db := dbConn(logger, t, config)
	defer db.Close()

	init := func(templateNamespace string, schemas []*schemas.Schema) {
		templateName := templateDBName(templateNamespace)
		name := pq.QuoteIdentifier(templateName)

		// We must first drop the template database because
		// migrations would not run on it if they had already ran,
		// even if the content of the migrations had changed during development.

		dbExec(t, db, `DROP DATABASE IF EXISTS `+name)
		dbExec(t, db, `CREATE DATABASE `+name+` TEMPLATE template0`)

		cfgCopy := *config
		cfgCopy.Path = "/" + templateName
		dbConn(logger, t, &cfgCopy, schemas...).Close()
	}

	init(templateNamespace, dbSchemas)
}

// templateDBName returns the name of the template database for the currently running package and namespace.
func templateDBName(templateNamespace string) string {
	parts := []string{
		"sourcegraph-test-template",
		wdHash(),
		templateNamespace,
	}

	return strings.Join(parts, "-")
}

// wdHash returns a hash of the current working directory.
// This is useful to get a stable identifier for the package running
// the tests.
func wdHash() string {
	h := fnv.New64()
	wd, _ := os.Getwd()
	h.Write([]byte(wd))
	return strconv.FormatUint(h.Sum64(), 10)
}

func dbConn(logger log.Logger, t testing.TB, cfg *url.URL, schemas ...*schemas.Schema) *sql.DB {
	t.Helper()
	db, err := connections.NewTestDB(t, logger, cfg.String(), schemas...)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") && os.Getenv("BAZEL_TEST") == "1" {
			t.Fatalf(`failed to connect to database %q: %s
PROTIP: Ensure the below is part of the go_test rule in BUILD.bazel
  tags = ["requires-network"]
See https://docs-legacy.sourcegraph.com/dev/background-information/bazel/faq#tests-fail-with-connection-refused`, cfg, err)
		}
		t.Fatalf("failed to connect to database %q: %s", cfg, err)
	}
	return db
}

func dbExec(t testing.TB, db *sql.DB, q string, args ...any) {
	t.Helper()
	_, err := db.Exec(q, args...)
	if err != nil {
		t.Errorf("failed to exec %q: %s", q, err)
	}
}

const killClientConnsQuery = `
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity WHERE datname = $1
`
