package dbtesting

import (
	"database/sql"
	"net/url"
	"os"
	"os/user"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// TimescaleDB returns a handle to the Code Insights TimescaleDB instance.
//
// The returned DB handle is initialized with a unique database just for the specified test, with
// all migrations applied.
func TimescaleDB(t testing.TB) (db *sql.DB, cleanup func()) {
	// Setup TimescaleDB for testing.
	if os.Getenv("CODEINSIGHTS_PGDATASOURCE") == "" {
		os.Setenv("CODEINSIGHTS_PGDATASOURCE", "postgres://postgres:password@127.0.0.1:5435/postgres")
	}
	username := ""
	if user, err := user.Current(); err == nil {
		username = user.Username
	}

	timescaleDSN := dbutil.PostgresDSN("codeinsights", username, os.Getenv)
	initConn, err := dbconn.NewRaw(timescaleDSN)
	if err != nil {
		t.Log("")
		t.Log("README: To run these tests you need to have the codeinsights TimescaleDB running:")
		t.Log("")
		t.Log("$ ./dev/codeinsights-db.sh &")
		t.Log("")
		t.Log("Or skip them with 'go test -short'")
		t.Log("")
		t.Logf("Failed to connect to codeinsights database: %s", err)
		if os.Getenv("CI") == "" {
			t.Skip()
		} else {
			t.FailNow()
		}
	}

	// Create database just for this test.
	dbname := "insights_test_" + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	_, err = initConn.Exec("DROP DATABASE IF EXISTS " + dbname + ";")
	if err != nil {
		t.Fatal("dropping test database", err)
	}
	_, err = initConn.Exec("CREATE DATABASE " + dbname + ";")
	if err != nil {
		t.Fatal("creating test database", err)
	}

	// Connect to the new DB.
	u, err := url.Parse(timescaleDSN)
	if err != nil {
		t.Fatal("parsing Timescale DSN", err)
	}
	u.Path = dbname
	timescaleDSN = u.String()
	db, err = dbconn.NewRaw(timescaleDSN)
	if err != nil {
		t.Fatal(err)
	}

	// Perform DB migrations.
	if err := dbconn.MigrateDB(db, dbconn.CodeInsights); err != nil {
		t.Fatalf("Failed to perform codeinsights database migration: %s", err)
	}
	cleanup = func() {
		if err := db.Close(); err != nil {
			t.Log(err)
		}
		defer initConn.Close()
		// It would be nice to cleanup by dropping the test DB we just created. But we can't:
		//
		// 	dropping test database ERROR: database "insights_test_testresolver_insights" is being accessed by other users (SQLSTATE 55006)
		//
		// This is because dbconn.MigrateDB leaks DB connections: https://github.com/sourcegraph/sourcegraph/pull/18033
		//
		// But, as you'll find there, we cannot have nice things because OF COURSE the fix somehow
		// passes all tests locally but not on our CI system. 💩💩💩
		//_, err = initConn.Exec(ctx, "DROP DATABASE "+dbname+";")
		//if err != nil {
		//	t.Fatal("dropping test database", err)
		//}
	}
	return db, cleanup
}
