package handlerutil

import (
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

// DebugMode is true if and only if the application is running in debug mode. In
// this mode, the application should display more verbose and informative errors
// in the UI. Debug should NEVER be true in production.
var DebugMode bool

func init() {
	DebugMode, _ = strconv.ParseBool(env.Get("DEBUG", "false", "debug mode"))
}
