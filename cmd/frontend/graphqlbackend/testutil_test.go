package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
)

func resetMocks() {
	backend.Mocks = backend.MockServices{}
}
