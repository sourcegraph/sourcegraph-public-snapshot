package backend

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"

	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

var Mocks MockServices

type MockServices struct {
	Annotations  MockAnnotations
	Async        MockAsync
	Defs         MockDefs
	MirrorRepos  MockMirrorRepos
	RepoStatuses MockRepoStatuses
	RepoTree     MockRepoTree
	Repos        MockRepos
	Search       MockSearch
}

// testContext creates a new context.Context for use by tests
func testContext() context.Context {
	localstore.Mocks = localstore.MockStores{}
	localstore.Graph = &localstore.Mocks.Graph
	Mocks = MockServices{}

	ctx := context.Background()
	ctx = authpkg.WithActor(ctx, &authpkg.Actor{UID: "1", Login: "test"})
	ctx = accesscontrol.WithInsecureSkip(ctx, true)
	_, ctx = opentracing.StartSpanFromContext(ctx, "dummy")

	return ctx
}
