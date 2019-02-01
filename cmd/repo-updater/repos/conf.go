package repos

import (
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func GetUpdateInterval() time.Duration {
	v := conf.Get().RepoListUpdateInterval
	if v == 0 { //  default to 1 minute
		v = 1
	}
	return time.Duration(v) * time.Minute
}
