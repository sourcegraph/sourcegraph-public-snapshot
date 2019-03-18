package repos

import (
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func getUpdateInterval() time.Duration {
	if v := conf.Get().RepoListUpdateInterval; v == 0 { //  default to 1 minute
		return 1 * time.Minute
	} else if v == -1 { // sentinel for zero
		return 0
	} else {
		return time.Duration(v) * time.Minute
	}
}
