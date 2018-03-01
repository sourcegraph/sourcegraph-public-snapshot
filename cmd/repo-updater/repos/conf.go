package repos

import (
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

func getUpdateInterval() time.Duration {
	v := time.Duration(conf.Get().RepoListUpdateInterval) * time.Minute
	if v == 0 {
		v = 1 * time.Minute // reasonable default
	}
	return v
}
