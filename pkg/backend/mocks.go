package backend

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
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
	db.Mocks = db.MockStores{}
	Mocks = MockServices{}

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "1"})
	_, ctx = opentracing.StartSpanFromContext(ctx, "dummy")

	return ctx
}
