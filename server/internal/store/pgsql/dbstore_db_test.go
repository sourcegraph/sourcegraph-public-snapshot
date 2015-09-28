// +build pgsqltest

package pgsql

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testdb"
)

// testContext constructs a new context that holds a temporary test DB
// handle. Call done() when done using it to release the DB handle to
// the pool so it can be used by other tests.
func testContext() (ctx context.Context, done func()) {
	dbh, done := testdb.NewHandle(&Schema)
	return NewContext(context.Background(), dbh), done
}
