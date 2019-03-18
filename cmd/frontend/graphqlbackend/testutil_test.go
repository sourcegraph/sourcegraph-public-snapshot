package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func init() {
	skipRefresh = true
}

func resetMocks() {
	db.Mocks = db.MockStores{}
	backend.Mocks = backend.MockServices{}
}
