package repos

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

var (
	reposListConf = conf.Get().ReposList
)

// RunRepositorySyncWorker runs the worker that syncs repositories from external code hosts to Sourcegraph
func RunRepositorySyncWorker(ctx context.Context) {
	configs := reposListConf
	if len(configs) == 0 {
		return
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
		for i, cfg := range configs {
			log15.Debug("RunRepositorySyncWorker:updateRepo", "repoURL", cfg.Url, "ith", i, "total", len(configs))
			err := updateRepo(ctx, cfg)
			if err != nil {
				log15.Warn("error updating repo", "path", cfg.Path, "error", err)
				continue
			}
		}
		time.Sleep(updateInterval)
	}
}

func updateRepo(ctx context.Context, repoConf schema.Repository) error {
	uri := api.RepoURI(repoConf.Path)
	repo, err := api.InternalClient.ReposCreateIfNotExists(ctx, api.RepoCreateOrUpdateRequest{RepoURI: uri, Enabled: true})
	if err != nil {
		return err
	}

	// Run a git fetch to kick-off an update or a clone if the repo doesn't already exist.
	cloned, err := gitserver.DefaultClient.IsRepoCloned(ctx, uri)
	if err != nil {
		return errors.Wrap(err, "error checking if repo cloned")
	}
	if !conf.Get().DisableAutoGitUpdates || !cloned {
		log15.Debug("fetching repos.list repo", "repo", uri, "cloned", cloned)
		err := gitserver.DefaultClient.EnqueueRepoUpdate(ctx, gitserver.Repo{Name: repo.URI, URL: repoConf.Url})
		if err != nil {
			return errors.Wrap(err, "error cloning repo")
		}
	}
	return nil
}

// GetExplicitlyConfiguredRepository reports information about a repository configured explicitly with "repos.list".
func GetExplicitlyConfiguredRepository(ctx context.Context, args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
	if args.Repo == "" {
		return nil, false, nil
	}

	repoNameLower := api.RepoURI(strings.ToLower(string(args.Repo)))
	for _, repo := range conf.Get().ReposList {
		if api.RepoURI(strings.ToLower(string(repo.Path))) == repoNameLower {
			repoInfo := &protocol.RepoInfo{
				URI:          api.RepoURI(repo.Path),
				ExternalRepo: nil,
				VCS:          protocol.VCSInfo{URL: repo.Url},
			}
			if repo.Links != nil {
				repoInfo.Links = &protocol.RepoLinks{
					Root:   repo.Links.Repository,
					Blob:   repo.Links.Blob,
					Tree:   repo.Links.Tree,
					Commit: repo.Links.Commit,
				}
			}
			return repoInfo, true, nil
		}
	}

	return nil, false, nil // not found
}
