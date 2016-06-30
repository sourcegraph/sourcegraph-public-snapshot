package auth

import (
	"os"
	"strconv"

	"golang.org/x/net/context"
)

var Debug bool

func init() {
	Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}

func DebugMode(ctx context.Context) bool {
	if env.Debug {
		return true
	}
	if ActorFromContext(ctx).Admin {
		return true
	}
	return false
}
