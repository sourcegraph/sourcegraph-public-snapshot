package repos

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	phabTaskRunning bool
	phabTaskMu      sync.Mutex
)

// RunGitoliteRepositorySyncWorker runs the worker that syncs repositories from gitolite hosts to Sourcegraph
//
// If there is a PhabricatorMetadataCommand set, every ten loops we will try to run that for every
// repo also.
func RunGitoliteRepositorySyncWorker(ctx context.Context) {
	phabricatorMetadataCounter := 0
	for {
		log15.Debug("RunGitoliteRepositorySyncWorker:GitoliteUpdateRepos")
		config, err := conf.GitoliteConfigs(context.Background())
		if err != nil {
			log15.Error("unable to fetch Gitolite configs", "err", err)
			continue
		}

		for _, gconf := range config {
			if err := gitoliteUpdateRepos(ctx, gconf, (phabricatorMetadataCounter%10) == 0); err != nil {
				log15.Error("error updating Gitolite repositories", "err", err, "prefix", gconf.Prefix)
			} else {
				log15.Debug("updated Gitolite repositories", "prefix", gconf.Prefix)
			}
		}
		// If you are looking at this line of code because this ran long enough that
		// an int wrapped, I'm sorry.
		phabricatorMetadataCounter++
		gitoliteUpdateTime.Set(float64(time.Now().Unix()))
		time.Sleep(getUpdateInterval())
	}
}

// GetGitoliteRepository returns dummy repo info about the repository, since we don't have the SSH keys in repo-updater
// to actually fetch metadata (and even if we did, it's unclear what metadata Gitolite has to offer beyond repository
// existence). We return a dummy response, because if we don't, callers will interpret the response as "repository not
// found".
func GetGitoliteRepository(ctx context.Context, args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
	config, err := conf.GitoliteConfigs(ctx)
	if err != nil {
		return nil, false, err
	}

	for _, c := range config {
		if strings.HasPrefix(string(args.Repo), c.Prefix) {
			return &protocol.RepoInfo{
				Name:         args.Repo,
				ExternalRepo: args.ExternalRepo,
			}, true, nil
		}
	}
	return nil, false, nil // not found
}

// tryUpdateGitolitePhabricatorMetadata attempts to update Phabricator metadata for a Gitolite-sourced repository, if it
// is appropriate to do so.
func tryUpdateGitolitePhabricatorMetadata(ctx context.Context, gconf *schema.GitoliteConnection, repos []string) {
	if gconf.PhabricatorMetadataCommand == "" {
		return
	}
	phabTaskMu.Lock()
	if phabTaskRunning {
		log15.Info("existing gitolite/phabricator repo task still running, skipping")
		phabTaskMu.Unlock()
		return
	}
	phabTaskRunning = true
	phabTaskMu.Unlock()
	for _, repoName := range repos {
		metadata, err := gitserver.DefaultClient.GetGitolitePhabricatorMetadata(ctx, gconf.Host, repoName)
		if err != nil {
			log15.Warn("could not fetch valid Phabricator metadata for Gitolite repository", "repo", repoName, "error", err)
			continue
		}
		if metadata.Callsign == "" {
			continue
		}
		if err := api.InternalClient.PhabricatorRepoCreate(ctx, api.RepoName(repoName), metadata.Callsign, gconf.Host); err != nil {
			log15.Warn("could not ensure Gitolite Phabricator mapping", "repo", repoName, "error", err)
		}
	}
	phabTaskMu.Lock()
	phabTaskRunning = false
	phabTaskMu.Unlock()
	log15.Info("updated gitolite/phabricator metadata for repos", "repos", len(repos))
}

// gitoliteUpdateRepos updates the repos associated with a specific
// Gitolite connection.
func gitoliteUpdateRepos(ctx context.Context, gconf *schema.GitoliteConnection, doPhabricator bool) error {
	// Get list of Gitolite repositories for this connection.
	rlist, err := gitserver.DefaultClient.ListGitolite(ctx, gconf.Host)
	if err != nil {
		return err
	}

	repoChan := make(chan repoCreateOrUpdateRequest)
	defer close(repoChan)
	go createEnableUpdateRepos(ctx, fmt.Sprintf("gitolite:%s", gconf.Prefix), repoChan)
	if doPhabricator && gconf.PhabricatorMetadataCommand != "" {
		go tryUpdateGitolitePhabricatorMetadata(ctx, gconf, rlist)
	}
	for _, entry := range rlist {
		// We don't have descriptions available for these. The old code didn't do that either.
		url := strings.Replace(entry, gconf.Prefix, gconf.Host+":", 1)
		repoChan <- repoCreateOrUpdateRequest{
			RepoCreateOrUpdateRequest: api.RepoCreateOrUpdateRequest{
				RepoName: api.RepoName(entry),
				Enabled:  true,
			},
			URL: url,
		}
	}
	return nil
}
