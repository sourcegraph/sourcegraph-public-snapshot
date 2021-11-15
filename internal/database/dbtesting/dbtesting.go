// Package dbtesting provides database test helpers.
package dbtesting

import (
	"context"
	"database/sql"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
)

// MockHashPassword if non-nil is used instead of database.hashPassword. This is useful
// when running tests since we can use a faster implementation.
var (
	MockHashPassword  func(password string) (sql.NullString, error)
	MockValidPassword func(hash, password string) bool
)

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

var (
	connectOnce sync.Once
	connectErr  error
)

// SetupGlobalTestDB creates a temporary test DB handle, sets
// `dbconn.Global` to it and setups other test configuration.
func SetupGlobalTestDB(t testing.TB) {
	useFastPasswordMocks()

	if testing.Short() {
		t.Skip()
	}

	connectOnce.Do(func() {
		connectErr = initTest()
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

	emptyDBPreserveSchema(t, dbconn.Global)
}

// GetDB calls SetupGlobalTestDB and returns dbconn.Global.
// It is meant to ease the migration away from dbconn.Global.
//
// New callers and callers actually wishing to migrate fully away from a global DB connection
// should use the new ../dbtest package instead of this one.
func GetDB(t testing.TB) *sql.DB {
	SetupGlobalTestDB(t)
	return dbconn.Global
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

// initTest creates a test database (dropping it if it already exists), and
// configures this package to use it.  It is called by integration tests (in a
// package init func) that need to use a real database.
func initTest() error {
	dbname, err := dbName()
	if err != nil {
		return err
	}

	if os.Getenv("TEST_SKIP_DROP_DB_BEFORE_TESTS") == "" {
		// When running the database-backcompat.sh tests, we need to *keep* the DB around because it has
		// the new schema produced by the new version. If we dropped the DB here, then we'd recreate
		// it at the OLD schema, which is not desirable because we need to run tests against the NEW
		// schema. Thus database-backcompat.sh sets TEST_SKIP_DROP_DB_BEFORE_TESTS=true.
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

	opts := dbconn.Opts{DSN: "dbname=" + dbname, DBName: dbname, AppName: "tests"}
	if err := dbconn.SetupGlobalConnection(opts); err != nil {
		return err
	}

	for _, database := range []*dbconn.Database{
		dbconn.Frontend,
		dbconn.CodeIntel,
	} {
		if err := dbconn.MigrateDB(dbconn.Global, database); err != nil {
			return err
		}
	}

	return nil
}

// dbName generates a unique name for the package currently being tested.
func dbName() (string, error) {
	pkg := testPkgName()
	if pkg == "" {
		return "", errors.New("dbtesting: could not detect test package")
	}

	// Postgres identifier limit is 64. Conservatively shorten name if bigger
	// than 32.
	if len(pkg) > 32 {
		pkg = fmt.Sprintf("%X-%s", crc32.ChecksumIEEE([]byte(pkg)), pkg[len(pkg)-32:])
	}

	return "sourcegraph-test-" + strings.ReplaceAll(pkg, "/", "-"), nil
}

// testPkgName finds the relative name of the sourcegraph package being tested
// by inspecting the call stack. If it fails, it returns an empty string.
func testPkgName() string {
	pc := make([]uintptr, 20)
	n := runtime.Callers(1, pc)
	if n == 0 {
		return ""
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	modulePrefix := "github.com/sourcegraph/sourcegraph/"
	pkg := ""

	var (
		frame runtime.Frame
		more  = true
	)

	// Look for last function name that looks like a sourcegraph test
	for more {
		frame, more = frames.Next()

		// Example name of a function we are looking for and the example pkg
		//
		//  github.com/sourcegraph/sourcegraph/cmd/frontend/backend.TestGetFirstServiceVersion
		//  =>
		//  cmd/frontend/backend

		testNameIdx := strings.Index(frame.Function, ".Test")
		if testNameIdx < 0 {
			continue
		}

		if !strings.HasPrefix(frame.Function, modulePrefix) {
			continue
		}

		pkg = frame.Function[len(modulePrefix):testNameIdx]
	}

	return pkg
}

// MockDB implements the dbutil.DB interface and is intended to be used
// in tests that require the database handle but never call it.
type MockDB struct {
	T *testing.T
}

func (db *MockDB) QueryContext(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error) {
	if db.T != nil {
		db.T.Fatal("mock db methods are not supposed to be called")
	}
	panic("mock db methods are not supposed to be called")
}

func (db *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if db.T != nil {
		db.T.Fatal("mock db methods are not supposed to be called")
	}
	panic("mock db methods are not supposed to be called")
}

func (db *MockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if db.T != nil {
		db.T.Fatal("mock db methods are not supposed to be called")
	}
	panic("mock db methods are not supposed to be called")
}
