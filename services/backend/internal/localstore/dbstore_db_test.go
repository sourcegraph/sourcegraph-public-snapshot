package localstore

import (
	"context"

	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store/mockstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/testdb"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	githubmock "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/mocks"
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

	appDBH, appDBDone := testdb.NewHandle("app", &AppSchema)
	graphDBH, graphDBDone := testdb.NewHandle("graph", &GraphSchema)

	dbCtx := WithAppDBH(ctx, appDBH)
	dbCtx = WithGraphDBH(dbCtx, graphDBH)
	return dbCtx, func() {
		appDBDone()
		graphDBDone()
	}
}

type mocks struct {
	mockstore.Stores
	mockstore.RepoVCS
	githubRepos githubmock.GitHubRepoGetter
}
