package graphqlbackend

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func init() {
	skipRefresh = true
}

func resetMocks() {
	db.Mocks = db.MockStores{}
	backend.Mocks = backend.MockServices{}
}
