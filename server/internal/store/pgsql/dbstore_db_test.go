// +build pgsqltest

package pgsql

import (
	"io/ioutil"
	"net/url"
	"os"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/server/internal/store/fs"
	"src.sourcegraph.com/sourcegraph/util/testdb"
)

// testContext constructs a new context that holds a temporary test DB
// handle and other test configuration. Call done() when done using it
// to release the DB handle to the pool so it can be used by other
// tests.
func testContext() (ctx context.Context, done func()) {
	ctx = context.Background()

	ctx = conf.WithURL(ctx, &url.URL{Scheme: "http", Host: "example.com"})

	reposDir, err := ioutil.TempDir("", "repos")
	if err != nil {
		panic("creating temp dir for repos: " + err.Error())
	}
	ctx = fs.WithReposVFS(ctx, reposDir)

	dbh, dbDone := testdb.NewHandle(&Schema)

	return NewContext(ctx, dbh), func() {
		dbDone()
		os.RemoveAll(reposDir)
	}
}
