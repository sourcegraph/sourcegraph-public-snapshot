package backend

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

var Mocks MockServices

type MockServices struct {
	Defs  MockDefs
	Pkgs  MockPkgs
	Orgs  MockOrgs
	Repos MockRepos
	Users MockUsers
}

// testContext creates a new context.Context for use by tests
func testContext() context.Context {
	localstore.Mocks = localstore.MockStores{}
	Mocks = MockServices{}

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "1", Login: "test"})
	_, ctx = opentracing.StartSpanFromContext(ctx, "dummy")

	return ctx
}
