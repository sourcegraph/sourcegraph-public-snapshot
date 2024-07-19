package own

import (
	"os"
	"strconv"
)

func IsEnabled() bool {
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_OWN")); v {
		return false
	}
	return true
}
