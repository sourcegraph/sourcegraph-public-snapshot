package externallink

import (
	"sourcegraph.com/cmd/frontend/backend"
	"sourcegraph.com/cmd/frontend/db"
	"sourcegraph.com/pkg/repoupdater"
)

func resetMocks() {
	repoupdater.MockRepoLookup = nil
	db.Mocks = db.MockStores{}
	backend.Mocks = backend.MockServices{}
}
