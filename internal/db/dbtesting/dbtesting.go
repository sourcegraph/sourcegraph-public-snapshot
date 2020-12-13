// Package dbtesting provides database test helpers.
package dbtesting

import (
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
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

// MockHashPassword if non-nil is used instead of db.hashPassword. This is useful
// when running tests since we can use a faster implementation.
var MockHashPassword func(password string) (sql.NullString, error)
var MockValidPassword func(hash, password string) bool

func useFastPasswordMocks() {
	// We can't care about security in tests, we care about speed.
	MockHashPassword = func(password string) (sql.NullString, error) {
		h := fnv.New64()
		_, _ = io.WriteString(h, password)
		return sql.NullString{Valid: true, String: strconv.FormatUint(h.Sum64(), 16)}, nil
	}
	MockValidPassword = func(hash, password string) bool {
		h := fnv.New64()
		_, _ = io.WriteString(h, password)
		return hash == strconv.FormatUint(h.Sum64(), 16)
	}
}

// BeforeTest functions are called before each test is run (by SetupGlobalTestDB).
var BeforeTest []func()

// DBNameSuffix must be set by DB test packages at init time to a value that is unique among all
// other such values used by other DB test packages. This is necessary to ensure the tests do not
// concurrently use the same DB (which would cause test failures).
var DBNameSuffix = "db"

var (
	connectOnce sync.Once
	connectErr  error
)

// SetupGlobalTestDB creates a temporary test DB handle, sets
// `dbconn.Global` to it and setups other test configuration.
//
// Callers (other than github.com/sourcegraph/sourcegraph/internal/db) must
// set a name in this package's DBNameSuffix var that is unique among all other
// test packages that call SetupGlobalTestDB, so that each package's
// tests run in separate DBs and do not conflict.
func SetupGlobalTestDB(t testing.TB) {
	SetupGlobalTestDBWithoutReset(t)

	emptyDBPreserveSchema(t, dbconn.Global)
}

// SetupGlobalTestDBWithoutReset creates a temporary test DB handle, sets
// `dbconn.Global` to it and setups other test configuration. It does NOT empty
// the DB if it already exists.
//
// Callers (other than github.com/sourcegraph/sourcegraph/internal/db) must
// set a name in this package's DBNameSuffix var that is unique among all other
// test packages that call SetupGlobalTestDBWithoutReset, so that each package's
// tests run in separate DBs and do not conflict.
func SetupGlobalTestDBWithoutReset(t testing.TB) {
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

	for _, f := range BeforeTest {
		f()
	}
}

func emptyDBPreserveSchema(t testing.TB, d *sql.DB) {
	_, err := d.Exec(`SELECT * FROM schema_migrations`)
	if err != nil {
		t.Fatalf("Table schema_migrations not found: %v", err)
	}

	var conds []string
	for _, migrationTable := range dbutil.MigrationTables {
		conds = append(conds, fmt.Sprintf("table_name != '%s'", migrationTable))
	}

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

	if err := dbconn.SetupGlobalConnection("dbname=" + dbname); err != nil {
		return err
	}

	for _, databaseName := range dbutil.DatabaseNames {
		if err := dbconn.MigrateDB(dbconn.Global, databaseName); err != nil {
			return err
		}
	}

	return nil
}

// FakeClock provides a controllable clock for use in tests.
type FakeClock struct {
	epoch time.Time
	step  time.Duration
	steps int
}

// NewFakeClock returns a FakeClock instance that starts telling time at the given epoch.
// Every invocation of Now adds step amount of time to the clock.
func NewFakeClock(epoch time.Time, step time.Duration) FakeClock {
	return FakeClock{epoch: epoch, step: step}
}

// Now returns the current fake time and advances the clock "step" amount of time.
func (c *FakeClock) Now() time.Time {
	c.steps++
	return c.Time(c.steps)
}

// Time tells the time at the given step from the provided epoch.
func (c FakeClock) Time(step int) time.Time {
	// We truncate to microsecond precision because Postgres' timestamptz type
	// doesn't handle nanoseconds.
	return c.epoch.Add(time.Duration(step) * c.step).UTC().Truncate(time.Microsecond)
}
