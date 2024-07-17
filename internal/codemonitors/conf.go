package codemonitors

import (
	"os"
	"strconv"
)

func IsEnabled() bool {
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_CODE_MONITORS")); v {
		return false
	}
	return true
}
