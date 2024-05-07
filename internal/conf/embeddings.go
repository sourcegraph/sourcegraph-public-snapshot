package conf

import (
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var allowEmbeddings, _ = strconv.ParseBool(env.Get(
	"SRC_INTERNAL_EMBEDDINGS_ALLOWED",
	"false",
	"Allow embeddings jobs and search to be enabled. NOTE: only intended for internal testing.",
))

// ForceAllowEmbeddings is true if we allow embeddings jobs and search to be enabled. It's only
// intended for internal evaluation and should never be enabled outside a development environment.
func ForceAllowEmbeddings() bool {
	return allowEmbeddings
}

type TB interface {
	Cleanup(func())
}

// MockForceAllowEmbeddings is used by tests to mock the result of ForceAllowEmbeddings.
func MockForceAllowEmbeddings(t TB, value bool) {
	orig := allowEmbeddings
	allowEmbeddings = value
	t.Cleanup(func() { allowEmbeddings = orig })
}
