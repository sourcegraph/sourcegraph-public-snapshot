package db

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
)

func init() {
	// We can't care about security in tests, we care about speed.
	mockHashPassword = func(password string) (sql.NullString, error) {
		h := fnv.New64()
		io.WriteString(h, password)
		return sql.NullString{Valid: true, String: strconv.FormatUint(h.Sum64(), 16)}, nil
	}
	mockValidPassword = func(hash, password string) bool {
		h := fnv.New64()
		io.WriteString(h, password)
		return hash == strconv.FormatUint(h.Sum64(), 16)
	}
}

func TestMigrations(t *testing.T) {
	if os.Getenv("SKIP_MIGRATION_TEST") != "" {
		t.Skip()
	}

	// get testing context to ensure we can connect to the DB
	_ = testContext(t)

	m := newMigrate(globalDB)
	// Run all down migrations then up migrations again to ensure there are no SQL errors.
	if err := m.Down(); err != nil {
		t.Errorf("error running down migrations: %s", err)
	}
	if err := doMigrateAndClose(m); err != nil {
		t.Errorf("error running up migrations: %s", err)
	}
}

func TestPassword(t *testing.T) {
	// By default we use fast mocks for our password in tests. This ensures
	// our actual implementation is correct.
	oldHash := mockHashPassword
	oldValid := mockValidPassword
	mockHashPassword = nil
	mockValidPassword = nil
	defer func() {
		mockHashPassword = oldHash
		mockValidPassword = oldValid
	}()

	h, err := hashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	if !validPassword(h.String, "correct-password") {
		t.Fatal("validPassword should of returned true")
	}
	if validPassword(h.String, "wrong-password") {
		t.Fatal("validPassword should of returned false")
	}
}

var (
	connectOnce sync.Once
	connectErr  error
)

// testContext constructs a new context that holds a temporary test DB
// handle and other test configuration.
func testContext(t *testing.T) context.Context {
	if testing.Short() {
		t.Skip()
	}

	connectOnce.Do(func() {
		connectErr = initTest("db")
	})
	if connectErr != nil {
		// only ignore connection errors if not on CI
		if os.Getenv("CI") == "" {
			t.Skip("Could not connect to DB", connectErr)
		}
		t.Fatal("Could not connect to DB", connectErr)
	}

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})

	Mocks = MockStores{}

	if err := emptyDBPreserveSchema(globalDB); err != nil {
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

// initTest creates a test database, named with the given suffix, if one does
// not already exist and configures this package to use it. It is called by
// integration tests (in a package init func) that need to use a real
// database.
func initTest(nameSuffix string) error {
	dbname := "sourcegraph-test-" + nameSuffix

	out, err := exec.Command("createdb", dbname).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "already exists") {
			log.Printf("DB %s exists already (run `dropdb %s` to delete and force re-creation)", dbname, dbname)
		} else {
			return errors.Errorf("createdb failed: %v\n%s", err, string(out))
		}
	}

	return ConnectToDB("dbname=" + dbname)
}
