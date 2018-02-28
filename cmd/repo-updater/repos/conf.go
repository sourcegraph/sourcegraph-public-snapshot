package repos

import (
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

var (
	updateInterval time.Duration
)

func init() {
	v := time.Duration(conf.GetTODO().RepoListUpdateInterval) * time.Minute
	if v == 0 {
		v = 1 * time.Minute // reasonable default
	}
	updateInterval = v
}
