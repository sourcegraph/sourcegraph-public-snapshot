package graphqlbackend

import (
	"sourcegraph.com/cmd/frontend/backend"
	"sourcegraph.com/cmd/frontend/db"
)

func resetMocks() {
	db.Mocks = db.MockStores{}
	backend.Mocks = backend.MockServices{}
}
