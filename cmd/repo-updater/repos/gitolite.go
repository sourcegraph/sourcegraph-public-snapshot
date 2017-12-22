package repos

import (
	"context"
	"log"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// RunRepositorySyncWorker runs the worker that syncs repositories from gitolite hosts to Sourcegraph
func RunGitoliteRepositorySyncWorker(ctx context.Context) error {
	if updateIntervalConf == 0 {
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

		time.Sleep(time.Duration(updateIntervalConf) * time.Minute)
	}
}
