package externallink

import (
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater"
)

func resetMocks() {
	repoupdater.MockRepoLookup = nil
	db.Mocks = db.MockStores{}
	backend.Mocks = backend.MockServices{}
}
