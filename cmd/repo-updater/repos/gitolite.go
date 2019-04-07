package repos

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	phabTaskRunning bool
	phabTaskMu      sync.Mutex
)

// RunGitoliteRepositorySyncWorker runs the worker that syncs repositories from gitolite hosts to Sourcegraph
//
// If there is a Phabricator set, every ten loops we will try to run that for every repo also.
func RunGitoliteRepositorySyncWorker(ctx context.Context) {
	phabricatorMetadataCounter := 0
	for {
		log15.Debug("RunGitoliteRepositorySyncWorker:GitoliteUpdateRepos")
		config, err := conf.GitoliteConfigs(context.Background())
		if err != nil {
			log15.Error("unable to fetch Gitolite configs", "err", err)
			time.Sleep(GetUpdateInterval())
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
		time.Sleep(GetUpdateInterval())
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
func tryUpdateGitolitePhabricatorMetadata(ctx context.Context, gconf *schema.GitoliteConnection, repoNames []api.RepoName) {
	if gconf.Phabricator == nil {
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
	for _, repoName := range repoNames {
		metadata, err := gitserver.DefaultClient.GetGitolitePhabricatorMetadata(ctx, gconf.Host, repoName)
		if err != nil {
			log15.Warn("could not fetch valid Phabricator metadata for Gitolite repository", "repo", repoName, "error", err)
			continue
		}
		if metadata.Callsign == "" {
			continue
		}
		if err := api.InternalClient.PhabricatorRepoCreate(ctx, repoName, metadata.Callsign, gconf.Phabricator.Url); err != nil {
			log15.Warn("could not ensure Gitolite Phabricator mapping", "repo", repoName, "error", err)
		}
	}
	phabTaskMu.Lock()
	phabTaskRunning = false
	phabTaskMu.Unlock()
	log15.Info("updated gitolite/phabricator metadata for repos", "repos", len(repoNames))
}

// gitoliteUpdateRepos updates the repos associated with a specific
// Gitolite connection.
func gitoliteUpdateRepos(ctx context.Context, gconf *schema.GitoliteConnection, doPhabricator bool) error {
	// Get list of Gitolite repositories for this connection.
	allRepos, err := gitserver.DefaultClient.ListGitolite(ctx, gconf.Host)
	if err != nil {
		return err
	}
	repos, err := filterBlacklist(gconf, allRepos)
	if err != nil {
		return err
	}

	repoChan := make(chan repoCreateOrUpdateRequest)
	defer close(repoChan)
	go createEnableUpdateRepos(ctx, fmt.Sprintf("gitolite:%s", gconf.Prefix), repoChan)
	if doPhabricator && gconf.Phabricator != nil {
		go tryUpdateGitolitePhabricatorMetadata(ctx, gconf, repoNames(gconf.Prefix, repos))
	}
	for _, gitoliteRepo := range repos {
		repoChan <- repoCreateOrUpdateRequest{
			RepoCreateOrUpdateRequest: api.RepoCreateOrUpdateRequest{
				RepoName:     reposource.GitoliteRepoName(gconf.Prefix, gitoliteRepo.Name),
				ExternalRepo: gitolite.ExternalRepoSpec(gitoliteRepo, gitolite.ServiceID(gconf.Host)),
				Enabled:      true,
			},
			URL: gitoliteRepo.URL,
		}
	}
	return nil
}

func filterBlacklist(gconf *schema.GitoliteConnection, allRepos []*gitolite.Repo) ([]*gitolite.Repo, error) {
	// filter out blacklist
	blacklist, err := blacklistRegexp(gconf.Blacklist)
	if err != nil {
		log15.Error("Invalid regexp for Gitolite blacklist", "expr", gconf.Blacklist, "err", err)
		return nil, err
	}
	blacklistCount := 0
	var repos []*gitolite.Repo
	for _, r := range allRepos {
		repoName := string(reposource.GitoliteRepoName(gconf.Prefix, r.Name))

		if strings.ContainsAny(repoName, "\\^$|()[]*?{},") || (blacklist != nil && blacklist.MatchString(repoName)) {
			blacklistCount++
			continue
		}
		repos = append(repos, r)
	}
	if blacklistCount > 0 {
		log15.Info("Excluded blacklisted Gitolite repositories", "num", blacklistCount)
	}
	return repos, nil
}

func blacklistRegexp(blacklistStr string) (*regexp.Regexp, error) {
	if blacklistStr == "" {
		return nil, nil
	}
	return regexp.Compile(blacklistStr)
}

func repoNames(prefix string, repos []*gitolite.Repo) []api.RepoName {
	names := make([]api.RepoName, 0, len(repos))
	for _, repo := range repos {
		names = append(names, reposource.GitoliteRepoName(prefix, repo.Name))
	}
	return names
}
