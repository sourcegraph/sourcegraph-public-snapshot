package bitbucketserver

import (
	"flag"
	"testing"

	"sourcegraph.com/pkg/db/dbtest"
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
