package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var Mocks MockServices

type MockServices struct {
	Repos MockRepos
}

// testContext creates a new context.Context for use by tests
func testContext() context.Context {
	Mocks = MockServices{}
	git.ResetMocks()

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	_, ctx = ot.StartSpanFromContext(ctx, "dummy")

	return ctx
}
