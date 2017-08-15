package graphqlbackend

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func init() {
	skipRefresh = true
}

func resetMocks() {
	localstore.Mocks = localstore.MockStores{}
	backend.Mocks = backend.MockServices{}
}
