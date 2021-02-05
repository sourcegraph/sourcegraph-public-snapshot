package dbtesting

import (
	"database/sql"
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
func TimescaleDB(t testing.TB) *sql.DB {
	// Setup TimescaleDB for testing.
	if os.Getenv("CODEINSIGHTS_PGDATASOURCE") == "" {
		os.Setenv("CODEINSIGHTS_PGDATASOURCE", "postgres://postgres:password@127.0.0.1:5435/postgres")
	}
	username := ""
	if user, err := user.Current(); err == nil {
		username = user.Username
	}
	timescaleDSN := dbutil.PostgresDSN("codeinsights", username, os.Getenv)
	db, err := dbconn.New(timescaleDSN, "insights-test-"+strings.Replace(t.Name(), "/", "_", -1))
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
			t.Fail()
		}
	}
	if err := dbconn.MigrateDB(db, dbconn.CodeInsights); err != nil {
		t.Fatalf("Failed to perform codeinsights database migration: %s", err)
	}
	return db
}
