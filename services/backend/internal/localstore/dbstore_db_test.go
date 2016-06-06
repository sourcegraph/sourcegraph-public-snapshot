package localstore

import (
	"net/url"

	"golang.org/x/net/context"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store/mockstore"
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
func testContext() (ctx context.Context, mock *mocks, done func()) {
	ctx = context.Background()

	ctx = conf.WithURL(ctx, &url.URL{Scheme: "http", Host: "example.com"})
	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: 1, Login: "test", Admin: true})
	ctx = accesscontrol.WithInsecureSkip(ctx, true)

	appDBH, appDBDone := testdb.NewHandle("app", &AppSchema)
	graphDBH, graphDBDone := testdb.NewHandle("graph", &GraphSchema)

	mock = &mocks{}
	ctx = store.WithStores(ctx, mock.Stores.Stores())
	ctx = store.WithRepoVCS(ctx, &mock.RepoVCS)

	dbCtx := WithAppDBH(ctx, appDBH)
	dbCtx = WithGraphDBH(dbCtx, graphDBH)
	return dbCtx, mock, func() {
		appDBDone()
		graphDBDone()
	}
}

type mocks struct {
	mockstore.Stores
	mockstore.RepoVCS
}
