package backend

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
)

var Mocks MockServices

type MockServices struct {
	Defs    MockDefs
	Repos   MockRepos
	Symbols MockSymbols

	Dependencies MockDependencies
	Packages     MockPackages
}

// testContext creates a new context.Context for use by tests
func testContext() context.Context {
	db.Mocks = db.MockStores{}
	Mocks = MockServices{}

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	_, ctx = opentracing.StartSpanFromContext(ctx, "dummy")

	return ctx
}
