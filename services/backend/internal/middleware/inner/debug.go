package inner

import (
	"os"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"

	"context"
)

var debug bool

func init() {
	debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}

// DebugMode returns true if and only if the application should make debug
// information such as error messages visible to clients of the gRPC service.
// DebugMode should NEVER return true in production.
func DebugMode(ctx context.Context) bool {
	if debug {
		return true
	}
	if auth.ActorFromContext(ctx).Admin {
		return true
	}
	return false
}
