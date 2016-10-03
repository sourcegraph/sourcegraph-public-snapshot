package localstore

import (
	"context"

	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/testdb"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
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

	ctx = authpkg.WithActor(ctx, &authpkg.Actor{UID: "1", Login: "test"})
	ctx = accesscontrol.WithInsecureSkip(ctx, true)

	Mocks = MockStores{}

	var appDBDone, graphDBDone func()
	globalAppDBH, appDBDone = testdb.NewHandle("app", &AppSchema)
	globalGraphDBH, graphDBDone = testdb.NewHandle("graph", &GraphSchema)

	return ctx, func() {
		appDBDone()
		graphDBDone()
	}
}
