package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

var Mocks MockServices

type MockServices struct {
	Repos MockRepos
}

// testContext creates a new context.Context for use by tests
func testContext() context.Context {
	Mocks = MockServices{}

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})

	return ctx
}
