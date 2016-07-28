package handlerutil

import (
	"os"
	"strconv"
)

// DebugMode is true if and only if the application is running in debug mode. In
// this mode, the application should display more verbose and informative errors
// in the UI. Debug should NEVER be true in production.
var DebugMode bool

func init() {
	DebugMode, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}
