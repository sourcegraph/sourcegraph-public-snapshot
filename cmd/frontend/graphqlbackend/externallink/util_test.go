package externallink

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
)

func resetMocks() {
	repoupdater.MockRepoLookup = nil
	database.Mocks = database.MockStores{}
	backend.Mocks = backend.MockServices{}
}
