package repos

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	updateIntervalConf = conf.Get().RepoListUpdateInterval

	logLevel = env.Get("SRC_LOG_LEVEL", "info", "upper log level to restrict log output to (dbug, dbug-dev, info, warn, error, crit)")
)
