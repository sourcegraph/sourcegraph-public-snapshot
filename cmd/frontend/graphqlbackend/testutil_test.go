package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
)

func resetMocks() {
	backend.Mocks = backend.MockServices{}
}
