package handlerutil

import (
	"os"
	"strconv"
)

var DebugMode bool

func init() {
	DebugMode, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}
