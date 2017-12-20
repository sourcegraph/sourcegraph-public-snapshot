package repos

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

var (
	reposListConf = conf.Get().ReposList
)

// RunRepositorySyncWorker runs the worker that syncs repositories from external code hosts to Sourcegraph
func RunRepositorySyncWorker(ctx context.Context) error {
	updateIntervalParsed, err := strconv.Atoi(updateIntervalEnv)
	if err != nil {
		return err
	}
	if updateIntervalParsed == 0 {
		return errors.New("Update interval is 0 (set REPOSITORY_SYNC_PERIOD to a non-zero value or omit it)")
	}
	updateInterval := time.Duration(updateIntervalParsed) * time.Second

	configs := reposListConf
	if len(configs) == 0 {
		return nil
	}

	for _, cfg := range configs {
		if cfg.Type == "" {
			cfg.Type = "git"
		}
		// We only support git repos at the moment.
		if cfg.Type != "git" {
			log15.Error("Error syncing repos, VCS type not supported", "type", cfg.Type, "repo", cfg.Path)
		}
	}
	for {
		for _, cfg := range configs {
			err := updateRepo(ctx, cfg)
			if err != nil {
				log15.Warn("error updating repo", "path", cfg.Path, "error", err)
				continue
			}
		}
		time.Sleep(updateInterval)
	}
}

func updateRepo(ctx context.Context, repoConf schema.ReposList) error {
	repo, err := sourcegraph.InternalClient.ReposCreateIfNotExists(ctx, repoConf.Path, "", false, false)
	if err != nil {
		return err
	}
	// Run a fetch kick-off an update or a clone if the repo doesn't already exist.
	cmd := gitserver.DefaultClient.Command("git", "fetch")
	cmd.Repo = repo
	err = cmd.Run(ctx)
	if err != nil {
		return errors.Wrap(err, "error cloning repo")
	}
	return nil
}
