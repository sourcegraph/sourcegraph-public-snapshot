package searchcontexts

import (
	"os"
	"strconv"
)

func IsEnabled() bool {
	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_SEARCH_CONTEXTS")); v {
		return false
	}
	return true
}
