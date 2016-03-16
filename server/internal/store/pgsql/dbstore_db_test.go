// +build pgsqltest

package pgsql

import (
	"net/url"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testdb"
)

func init() {
	skipFS = true
}

// testContext constructs a new context that holds a temporary test DB
// handle and other test configuration. Call done() when done using it
// to release the DB handle to the pool so it can be used by other
// tests.
func testContext() (ctx context.Context, done func()) {
	ctx = context.Background()

	ctx = conf.WithURL(ctx, &url.URL{Scheme: "http", Host: "example.com"})

	dbh, dbDone := testdb.NewHandle(&Schema)

	return NewContext(ctx, dbh), func() {
		dbDone()
	}
}
