package graphqlbackend

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func resetMocks() {
	localstore.Mocks = localstore.MockStores{}
	backend.Mocks = backend.MockServices{}
}
