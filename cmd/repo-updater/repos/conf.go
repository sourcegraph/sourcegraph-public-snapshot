package repos

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func GetUpdateInterval() time.Duration {
	v := conf.Get().RepoListUpdateInterval
	if v == 0 { //  default to 1 minute
		v = 1
	}
	return time.Duration(v) * time.Minute
}

func GetSyncConcurrency() int {
	v := conf.Get().RepoSyncConcurrency
	if v == 0 { //  default to 1 minute
		v = 3
	}
	return v
}
