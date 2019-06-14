package bitbucketserver

import (
	"context"
	"database/sql"
	"flag"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbtest"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := dbtest.NewDB(t, *dsn)
	defer cleanup()

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"Store", testStore(db)},
		{"Provider/RepoPerms", testProviderRepoPerms(db)},
	} {
		t.Run(tc.name, tc.test)
	}
}

func transact(db *sql.DB, test func(*sql.Tx)) func(*testing.T) {
	return func(t *testing.T) {
		tx, err := db.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}

		defer tx.Rollback()

		test(tx)
	}
}
