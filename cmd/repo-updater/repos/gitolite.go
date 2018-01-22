package repos

import (
	"context"
	"log"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// RunGitoliteRepositorySyncWorker runs the worker that syncs repositories from gitolite hosts to Sourcegraph
func RunGitoliteRepositorySyncWorker(ctx context.Context) error {
	for {
		if err := api.InternalClient.GitoliteUpdateRepos(context.Background()); err != nil {
			log.Println(err)
		} else {
			log15.Debug("updated Gitolite repos")
		}

		time.Sleep(updateInterval)
	}
}
