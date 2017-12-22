package main

import (
	"context"
	"log"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	logLevel = env.Get("SRC_LOG_LEVEL", "info", "upper log level to restrict log output to (dbug, dbug-dev, info, warn, error, crit)")
)

func main() {
	interval := conf.Get().RepoListUpdateInterval
	if interval == 0 {
		log.Println("REPO_LIST_UPDATE_INTERVAL not set, not doing any updates")
		select {}
	}

	// Filter log output by level.
	lvl, err := log15.LvlFromString(logLevel)
	if err != nil {
		log.Fatalf("could not parse log level: %v", err)
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, log15.StderrHandler))

	for {
		if err := sourcegraph.InternalClient.GitoliteUpdateRepos(context.Background()); err != nil {
			log.Println(err)
		} else {
			log15.Debug("updated Gitolite repos")
		}

		time.Sleep(time.Duration(interval) * time.Minute)
	}
}
