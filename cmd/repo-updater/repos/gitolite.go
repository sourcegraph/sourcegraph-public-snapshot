package repos

import (
	"context"
	"log"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

// RunGitoliteRepositorySyncWorker runs the worker that syncs repositories from gitolite hosts to Sourcegraph
func RunGitoliteRepositorySyncWorker(ctx context.Context) error {
	// Filter log output by level.
	lvl, err := log15.LvlFromString(env.LogLevel)
	if err != nil {
		log.Fatalf("could not parse log level: %v", err)
	}
	log15.Root().SetHandler(log15.LvlFilterHandler(lvl, log15.StderrHandler))

	for {
		if err := api.InternalClient.GitoliteUpdateRepos(context.Background()); err != nil {
			log.Println(err)
		} else {
			log15.Debug("updated Gitolite repos")
		}

		time.Sleep(updateInterval)
	}
}
