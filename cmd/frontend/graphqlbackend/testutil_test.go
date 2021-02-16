package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func resetMocks() {
	database.Mocks = database.MockStores{}
	backend.Mocks = backend.MockServices{}
}
