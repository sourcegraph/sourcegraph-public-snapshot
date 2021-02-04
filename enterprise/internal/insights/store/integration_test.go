package store

import (
	"context"
	"database/sql"
	"os"
	"os/user"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	getTimescaleDB := func(t testing.TB) *sql.DB {
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

	t.Run("Integration", func(t *testing.T) {
		ctx := context.Background()
		clock := timeutil.Now
		store := NewWithClock(getTimescaleDB(t), clock)
		t.Run("Insights", func(t *testing.T) { testInsights(t, ctx, store, clock) })
	})
}
