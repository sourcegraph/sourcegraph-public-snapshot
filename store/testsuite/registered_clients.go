package testsuite

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// RegisteredClients_Delete_nonexistent tests the behavior of
// RegisteredClients.Delete when called with a nonexistent client ID.
func RegisteredClients_Delete_nonexistent(ctx context.Context, t *testing.T, s store.RegisteredClients) {
	if err := s.Delete(ctx, sourcegraph.RegisteredClientSpec{ID: "doesntexist"}); !isRegisteredClientNotFound(err) {
		t.Fatal(err)
	}
}

func isRegisteredClientNotFound(err error) bool {
	_, ok := err.(*store.RegisteredClientNotFoundError)
	return ok
}
