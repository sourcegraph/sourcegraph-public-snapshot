package repos

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
)

// ConfRepoListUpdateInterval returns the repository list update interval.
//
// If the RepoListUpdateInterval site configuration setting is 0, it defaults to:
//
// - 15 seconds for app deployments (to speed up adding repos during setup)
// - 1 minute otherwise
func ConfRepoListUpdateInterval() time.Duration {
	v := conf.Get().RepoListUpdateInterval
	if v == 0 { //  default to 1 minute
		if deploy.IsApp() {
			return time.Second * 15 // 15 seconds for app deployments
		}
		v = 1
	}
	return time.Duration(v) * time.Minute
}

func ConfRepoConcurrentExternalServiceSyncers() int {
	v := conf.Get().RepoConcurrentExternalServiceSyncers
	if v <= 0 {
		return 3
	}
	return v
}
