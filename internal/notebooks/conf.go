package notebooks

import (
	"os"
	"strconv"
)

func IsEnabled() bool {
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_NOTEBOOKS")); v {
		return false
	}
	return true
}
