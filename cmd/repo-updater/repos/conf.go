package repos

import (
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	updateInterval time.Duration
	logLevel       = env.Get("SRC_LOG_LEVEL", "info", "upper log level to restrict log output to (dbug, dbug-dev, info, warn, error, crit)")
)

func init() {
	v := time.Duration(conf.Get().RepoListUpdateInterval) * time.Minute
	if v == 0 {
		v = 1 * time.Minute // reasonable default
	}
	updateInterval = v
}
