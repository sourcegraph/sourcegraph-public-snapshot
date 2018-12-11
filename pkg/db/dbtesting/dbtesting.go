// Package dbtesting provides database test helpers.
package dbtesting

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// MockHashPassword if non-nil is used instead of db.hashPassword. This is useful
// when running tests since we can use a faster implementation.
var MockHashPassword func(password string) (sql.NullString, error)
var MockValidPassword func(hash, password string) bool

func useFastPasswordMocks() {
	// We can't care about security in tests, we care about speed.
	MockHashPassword = func(password string) (sql.NullString, error) {
		h := fnv.New64()
		io.WriteString(h, password)
		return sql.NullString{Valid: true, String: strconv.FormatUint(h.Sum64(), 16)}, nil
	}
	MockValidPassword = func(hash, password string) bool {
		h := fnv.New64()
		io.WriteString(h, password)
		return hash == strconv.FormatUint(h.Sum64(), 16)
	}
}

// BeforeTest functions are called before each test is run (by TestContext).
var BeforeTest []func()

// DBNameSuffix must be set by DB test packages at init time to a value that is unique among all
// other such values used by other DB test packages. This is necessary to ensure the tests do not
// concurrently use the same DB (which would cause test failures).
var DBNameSuffix = "db"

var (
	connectOnce sync.Once
	connectErr  error
)

// TestContext constructs a new context that holds a temporary test DB
// handle and other test configuration.
//
// Callers (other than github.com/sourcegraph/sourcegraph/cmd/frontend/db) must set a name in this
// package's DBNameSuffix var that is unique among all other test packages that call TestContext, so
// that each package's tests run in separate DBs and do not conflict.
func TestContext(t *testing.T) context.Context {
	useFastPasswordMocks()

	if testing.Short() {
		t.Skip()
	}

	connectOnce.Do(func() {
		connectErr = initTest(DBNameSuffix)
	})
	if connectErr != nil {
		// only ignore connection errors if not on CI
		if os.Getenv("CI") == "" {
			t.Skip("Could not connect to DB", connectErr)
		}
		t.Fatal("Could not connect to DB", connectErr)
	}

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	for _, f := range BeforeTest {
		f()
	}

	if err := emptyDBPreserveSchema(dbconn.Global); err != nil {
		log.Fatal(err)
	}

	return ctx
}

func emptyDBPreserveSchema(d *sql.DB) error {
	_, err := d.Exec(`SELECT * FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("Table schema_migrations not found: %v", err)
	}
	return truncateDB(d)
}

func truncateDB(d *sql.DB) error {
	rows, err := d.Query("SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' AND table_name != 'schema_migrations'")
	if err != nil {
		return err
	}
	var tables []string
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	log.Printf("Truncating all %d tables", len(tables))
	_, err = d.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY")
	return err
}

// initTest creates a test database, named with the given suffix
// (dropping it if it already exists), and configures this package to use it.
// It is called by integration tests (in a package init func) that need to use
// a real database.
func initTest(nameSuffix string) error {
	dbname := "sourcegraph-test-" + nameSuffix

	if os.Getenv("TEST_SKIP_DROP_DB_BEFORE_TESTS") == "" {
		// When running the db-backcompat.sh tests, we need to *keep* the DB around because it has
		// the new schema produced by the new version. If we dropped the DB here, then we'd recreate
		// it at the OLD schema, which is not desirable because we need to run tests against the NEW
		// schema. Thus db-backcompat.sh sets TEST_SKIP_DROP_DB_BEFORE_TESTS=true.
		out, err := exec.Command("dropdb", "--if-exists", dbname).CombinedOutput()
		if err != nil {
			return errors.Errorf("dropdb --if-exists failed: %v\n%s", err, string(out))
		}
	}

	out, err := exec.Command("createdb", dbname).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "already exists") {
			log.Printf("DB %s exists already (run `dropdb %s` to delete and force re-creation)", dbname, dbname)
		} else {
			return errors.Errorf("createdb failed: %v\n%s", err, string(out))
		}
	}

	return dbconn.ConnectToDB("dbname=" + dbname)
}
