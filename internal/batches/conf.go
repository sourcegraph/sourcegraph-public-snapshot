package batches

import (
	"os"
	"strconv"
)

func IsEnabled() bool {
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_BATCH_CHANGES")); v {
		return false
	}
	return true
}
