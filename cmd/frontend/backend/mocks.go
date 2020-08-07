package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var Mocks, emptyMocks struct {
	Repos                MockRepos
	MakePasswordResetURL func(context.Context, int32) (string, error)
}

// testContext creates a new context.Context for use by tests
func testContext() context.Context {
	db.Mocks = db.MockStores{}
	Mocks = emptyMocks
	git.ResetMocks()

	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})
	_, ctx = ot.StartSpanFromContext(ctx, "dummy")

	return ctx
}

// ResetMocks clears the mock functions set on Mocks (so that subsequent tests don't inadvertently
// use them).
func ResetMocks() {
	Mocks = emptyMocks
}
