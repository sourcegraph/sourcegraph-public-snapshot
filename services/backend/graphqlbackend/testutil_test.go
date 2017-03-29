package graphqlbackend

import (
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/localstore"
)

func resetMocks() {
	localstore.Mocks = localstore.MockStores{}
	backend.Mocks = backend.MockServices{}
}
