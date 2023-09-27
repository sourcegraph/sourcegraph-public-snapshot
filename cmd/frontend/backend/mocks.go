pbckbge bbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
)

vbr Mocks MockServices

type MockServices struct {
	Repos MockRepos
}

// testContext crebtes b new context.Context for use by tests
func testContext() context.Context {
	Mocks = MockServices{}

	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1})

	return ctx
}
