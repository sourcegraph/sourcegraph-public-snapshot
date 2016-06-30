package inner

import (
	"os"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"

	"golang.org/x/net/context"
)

var debug bool

func init() {
	debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}

// DebugMode returns true if and only if the application should make debug
// information such as error messages visible to clients of the gRPC service.
func DebugMode(ctx context.Context) bool {
	if debug {
		return true
	}
	if auth.ActorFromContext(ctx).Admin {
		return true
	}
	return false
}
