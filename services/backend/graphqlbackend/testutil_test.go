package graphqlbackend

import (
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

func resetMocks() {
	localstore.Mocks = localstore.MockStores{}
	localstore.Graph = &localstore.Mocks.Graph
	backend.Mocks = backend.MockServices{}
}
