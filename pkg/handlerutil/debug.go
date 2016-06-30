package handlerutil

import (
	"os"
	"strconv"

	"golang.org/x/net/context"
)

var debug bool

func init() {
	debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}

// DebugMode returns true if and only if the application should make debug
// information such as error messages visible in the UI. This is similar to the
// DebugMode at the gRPC layer, but it doesn't check for user admin privileges,
// because the actor should not be set in the httpapi layer.
func DebugMode(ctx context.Context) bool {
	if debug {
		return true
	}
	return false
}
