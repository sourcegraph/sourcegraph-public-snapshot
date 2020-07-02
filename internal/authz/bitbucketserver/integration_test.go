package bitbucketserver

import (
	"flag"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db := dbtest.NewDB(t, *dsn)

	cli, save := newClient(t, "BitbucketServer")
	defer save()

	f := newFixtures()
	f.load(t, cli)

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"Store", testStore(db)},
		{"Provider/RepoPerms", testProviderRepoPerms(db, f, cli)},
		{"Provider/FetchAccount", testProviderFetchAccount(f, cli)},
		{"Provider/FetchUserPerms", testProviderFetchUserPerms(f, cli)},
		{"Provider/FetchRepoPerms", testProviderFetchRepoPerms(f, cli)},
	} {
		t.Run(tc.name, tc.test)
	}
}
