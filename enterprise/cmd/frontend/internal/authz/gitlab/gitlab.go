// Package gitlab contains an authorization provider for GitLab that uses GitLab OAuth
// authenetication.
package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var _ authz.Provider = ((*GitLabOAuthAuthzProvider)(nil))

type GitLabOAuthAuthzProvider struct {
	clientProvider *gitlab.ClientProvider
	clientURL      *url.URL
	codeHost       *gitlab.CodeHost
	cache          cache
	cacheTTL       time.Duration
}

type GitLabOAuthAuthzProviderOp struct {
	// BaseURL is the URL of the GitLab instance.
	BaseURL *url.URL

	// CacheTTL is the TTL of cached permissions lists from the GitLab API.
	CacheTTL time.Duration

	// MockCache, if non-nil, replaces the default Redis-based cache with the supplied cache mock.
	// Should only be used in tests.
	MockCache cache
}

func NewProvider(op GitLabOAuthAuthzProviderOp) *GitLabOAuthAuthzProvider {
	p := &GitLabOAuthAuthzProvider{
		clientProvider: gitlab.NewClientProvider(op.BaseURL, nil),
		clientURL:      op.BaseURL,
		codeHost:       gitlab.NewCodeHost(op.BaseURL),
		cache:          op.MockCache,
		cacheTTL:       op.CacheTTL,
	}
	if p.cache == nil {
		p.cache = rcache.NewWithTTL(fmt.Sprintf("gitlabAuthz:%s", op.BaseURL.String()), int(math.Ceil(op.CacheTTL.Seconds())))
	}
	return p
}

func (p *GitLabOAuthAuthzProvider) Validate() (problems []string) {
	return nil
}

func (p *GitLabOAuthAuthzProvider) ServiceID() string {
	return p.codeHost.ServiceID()
}

func (p *GitLabOAuthAuthzProvider) ServiceType() string {
	return p.codeHost.ServiceType()
}

func (p *GitLabOAuthAuthzProvider) RepoPerms(ctx context.Context, account *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[api.RepoName]map[authz.Perm]bool, error) {
	accountID := "" // empty means public / unauthenticated to the code host
	if account != nil && account.ServiceID == p.codeHost.ServiceID() && account.ServiceType == p.codeHost.ServiceType() {
		accountID = account.AccountID
	}

	myRepos, _ := p.Repos(ctx, repos)
	var accessibleRepos map[int]struct{}
	if r, exists := p.getCachedAccessList(accountID); exists {
		accessibleRepos = r
	} else {
		var accessToken string
		if account != nil {
			_, tok, err := gitlab.GetExternalAccountData(&account.ExternalAccountData)
			if err != nil {
				return nil, err
			}
			accessToken = tok.AccessToken
		}

		var err error
		accessibleRepos, err = p.fetchUserAccessList(ctx, accountID, accessToken)
		if err != nil {
			return nil, err
		}

		accessibleReposB, err := json.Marshal(cacheVal{
			ProjIDs: accessibleRepos,
			TTL:     p.cacheTTL,
		})
		if err != nil {
			return nil, err
		}
		p.cache.Set(accountID, accessibleReposB)
	}

	perms := make(map[api.RepoName]map[authz.Perm]bool)
	for repo := range myRepos {
		perms[repo.RepoName] = map[authz.Perm]bool{}

		projID, err := strconv.Atoi(repo.ExternalRepoSpec.ID)
		if err != nil {
			log15.Warn("couldn't parse GitLab proj ID as int while computing permissions", "id", repo.ExternalRepoSpec.ID)
			continue
		}
		_, isAccessible := accessibleRepos[projID]
		if !isAccessible {
			continue
		}
		perms[repo.RepoName][authz.Read] = true
	}

	return perms, nil
}

func (p *GitLabOAuthAuthzProvider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	return authz.GetCodeHostRepos(p.codeHost, repos)
}

func (p *GitLabOAuthAuthzProvider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.ExternalAccount) (mine *extsvc.ExternalAccount, err error) {
	return nil, nil
}

// getCachedAccessList returns the list of repositories accessible to a user from the cache and
// whether the cache entry exists.
func (p *GitLabOAuthAuthzProvider) getCachedAccessList(accountID string) (map[int]struct{}, bool) {
	// TODO(beyang): trigger best-effort fetch in background if ttl is getting close (but avoid dup refetches)

	cachedReposB, exists := p.cache.Get(accountID)
	if !exists {
		return nil, false
	}
	var r cacheVal
	if err := json.Unmarshal(cachedReposB, &r); err != nil || r.TTL == 0 || r.TTL > p.cacheTTL {
		if err != nil {
			log15.Warn("Failed to unmarshal repo perm cache entry", "err", err.Error())
		}
		p.cache.Delete(accountID)
		return nil, false
	}
	return r.ProjIDs, true
}

// fetchUserAccessList fetches the list of project IDs that are readable to a user from the GitLab API.
func (p *GitLabOAuthAuthzProvider) fetchUserAccessList(ctx context.Context, glUserID, oauthTok string) (map[int]struct{}, error) {
	projIDs := make(map[int]struct{})
	var iters = 0
	var pageURL = fmt.Sprintf("projects?%s", url.Values(map[string][]string{"per_page": {"100"}}).Encode())
	for {
		if iters >= 100 && iters%100 == 0 {
			log15.Warn("Excessively many GitLab API requests to fetch complete user authz list", "iters", iters, "gitlabUserID", glUserID, "host", p.clientURL.String())
		}

		projs, nextPageURL, err := p.clientProvider.GetOAuthClient(oauthTok).ListProjects(ctx, pageURL)
		if err != nil {
			return nil, err
		}
		for _, proj := range projs {
			projIDs[proj.ID] = struct{}{}
		}

		if nextPageURL == nil {
			break
		}
		pageURL = *nextPageURL
		iters++
	}
	return projIDs, nil
}
