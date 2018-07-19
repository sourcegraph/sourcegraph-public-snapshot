package repos

import (
	"context"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

// RunGitoliteRepositorySyncWorker runs the worker that syncs repositories from gitolite hosts to Sourcegraph
func RunGitoliteRepositorySyncWorker(ctx context.Context) {
	for {
		start := time.Now()
		log15.Debug("RunGitoliteRepositorySyncWorker:GitoliteUpdateRepos")
		if err := api.InternalClient.GitoliteUpdateRepos(context.Background()); err != nil {
			log15.Error("Error updating Gitolite repositories.", "err", err)
		} else {
			log15.Info("Updated Gitolite repositories.", "duration", time.Now().Sub(start))
		}
		gitoliteUpdateTime.Set(float64(time.Now().Unix()))
		time.Sleep(getUpdateInterval())
	}
}

// GetGitoliteRepository returns dummy repo info about the repository, since we don't have the SSH keys in repo-updater
// to actually fetch metadata (and even if we did, it's unclear what metadata Gitolite has to offer beyond repository
// existence). We return a dummy response, because if we don't, callers will interpret the response as "repository not
// found".
func GetGitoliteRepository(ctx context.Context, args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
	for _, c := range conf.Get().Gitolite {
		if strings.HasPrefix(string(args.Repo), c.Prefix) {
			return &protocol.RepoInfo{
				URI:          args.Repo,
				ExternalRepo: args.ExternalRepo,
			}, true, nil
		}
	}
	return nil, false, nil // not found
}
