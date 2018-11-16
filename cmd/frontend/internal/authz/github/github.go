package github

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
)

// Provider implements authz.Provider for GitHub repository permissions.
type Provider struct {
	client   *github.Client
	codeHost *github.CodeHost
	cacheTTL time.Duration
	cache    pcache
}

func NewProvider(githubURL *url.URL, baseToken string, cacheTTL time.Duration, mockCache pcache) *Provider {
	apiURL, _ := github.APIRoot(githubURL)
	client := github.NewClient(apiURL, baseToken, nil)

	p := &Provider{
		codeHost: github.NewCodeHost(githubURL),
		client:   client,
		cache:    mockCache,
		cacheTTL: cacheTTL,
	}
	// Note: this will use the same underlying Redis instance and key namespace for every instance
	// of Provider.  This is by design, so that different instances, even in different processes,
	// will share cache entries.
	if p.cache == nil {
		p.cache = rcache.NewWithTTL(fmt.Sprintf("githubAuthz:%s", githubURL.String()), int(math.Ceil(cacheTTL.Seconds())))
	}
	return p
}

var _ authz.Provider = ((*Provider)(nil))

// Repos implements the authz.Provider interface.
func (p *Provider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	return authz.GetCodeHostRepos(p.codeHost, repos)
}

// RepoPerms implements the authz.Provider interface.
//
// It computes permissions by keeping track of two classes of info:
// * Whether a given repository is public
// * Whether a given user can access a given repository
//
// For each repo in the input set, it first checks if it is public. If not, it checks if the given
// user has access to it. It caches both the "is repo public" and "user can access this repo" values
// in Redis. If missing from the cache, it makes a GitHub API request to determine the value. It
// makes a separate API request for each repo (this can later be optimized if necessary).
func (p *Provider) RepoPerms(ctx context.Context, userAccount *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[api.RepoName]map[authz.Perm]bool, error) {
	repos, _ = p.Repos(ctx, repos)
	if len(repos) == 0 {
		return nil, nil
	}

	explicitRepos, err := p.userRepos(ctx, userAccount, repos)
	if err != nil {
		return nil, err
	}

	perms := make(map[api.RepoName]map[authz.Perm]bool) // permissions to return
	var nonExplicitRepos map[authz.Repo]struct{}
	if explicitRepos == nil {
		nonExplicitRepos = repos
	} else {
		// repos to which user doesn't have explicit access
		nonExplicitRepos = map[authz.Repo]struct{}{}
		for repo := range repos {
			if hasAccess, ok := explicitRepos[repo.ExternalRepoSpec.ID]; ok {
				perms[repo.RepoName] = map[authz.Perm]bool{authz.Read: hasAccess}
			} else {
				nonExplicitRepos[repo] = struct{}{}
			}
		}
	}

	if len(nonExplicitRepos) > 0 {
		publicRepos, err := p.publicRepos(ctx, nonExplicitRepos)
		if err != nil {
			return nil, err
		}
		if publicRepos != nil {
			for repo := range nonExplicitRepos {
				if publicRepos[repo.ExternalRepoSpec.ID] {
					perms[repo.RepoName] = map[authz.Perm]bool{authz.Read: true}
				}
			}
		}
	}
	return perms, nil
}

// publicRepos accepts a set of repositories and returns a map from repository external ID (the
// GitHub repository GraphQL ID) to true/false indicating whether the repository is public or
// private. It consults and updates the cache.
func (p *Provider) publicRepos(ctx context.Context, repos map[authz.Repo]struct{}) (map[string]bool, error) {
	cachedIsPublic, err := p.getCachedPublicRepos(ctx, repos)
	if err != nil {
		return nil, err
	}
	if len(cachedIsPublic) == len(repos) {
		return cachedIsPublic, nil
	}

	missing := make(map[string]struct{})
	for r := range repos {
		if _, ok := cachedIsPublic[r.ExternalRepoSpec.ID]; !ok {
			missing[r.ExternalRepoSpec.ID] = struct{}{}
		}
	}

	missingIsPublic, err := p.fetchPublicRepos(ctx, missing)
	if err != nil {
		return nil, err
	}
	p.setCachedPublicRepos(ctx, missingIsPublic)

	for k, v := range missingIsPublic {
		cachedIsPublic[k] = v
	}
	return cachedIsPublic, nil
}

// setCachedPublicRepos updates the cache with a map from GitHub repo ID to true/false indicating
// whether the repo is public or private. The GitHub repo ID is the GraphQL API ID ("repository node
// ID").
//
// Internally, it sets a separate cache key for each repo ID.
func (p *Provider) setCachedPublicRepos(ctx context.Context, isPublic map[string]bool) error {
	setArgs := make([][2]string, 0, len(isPublic))
	for k, v := range isPublic {
		key := publicRepoCacheKey(k)
		val, err := json.Marshal(publicRepoCacheVal{
			Public: v,
			TTL:    p.cacheTTL,
		})
		if err != nil {
			return err
		}
		setArgs = append(setArgs, [2]string{key, string(val)})
	}
	p.cache.SetMulti(setArgs...)
	return nil
}

// getCachedPublicRepos accepts a set of repos and returns a map from repo ID to true/false
// indicating whether the repo is public or private. The returned map may be incomplete (i.e., not
// every input repo may be represented in the key set) due to cache incompleteness.
func (p *Provider) getCachedPublicRepos(ctx context.Context, repos map[authz.Repo]struct{}) (isPublic map[string]bool, err error) {
	if len(repos) == 0 {
		return nil, nil
	}
	isPublic = make(map[string]bool)
	repoList := make([]string, 0, len(repos))
	getArgs := make([]string, 0, len(repos))
	for r := range repos {
		getArgs = append(getArgs, publicRepoCacheKey(r.ExternalRepoSpec.ID))
		repoList = append(repoList, r.ExternalRepoSpec.ID)
	}
	vals := p.cache.GetMulti(getArgs...)
	for i, v := range vals {
		if len(v) == 0 {
			continue
		}
		var val publicRepoCacheVal
		if err := json.Unmarshal(v, &val); err != nil {
			return nil, err
		}
		isPublic[repoList[i]] = val.Public
	}

	return isPublic, nil
}

// fetchPublicRepos returns a map from GitHub repository ID (the GraphQL repo node ID) to true/false
// indicating whether a repository is public (true) or private (false).
func (p *Provider) fetchPublicRepos(ctx context.Context, repos map[string]struct{}) (map[string]bool, error) {
	isPublic := make(map[string]bool)
	for ghRepoID := range repos {
		ghRepo, err := p.client.GetRepositoryByNodeID(ctx, "", ghRepoID)
		if err == github.ErrNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		isPublic[ghRepoID] = !ghRepo.IsPrivate
	}
	return isPublic, nil
}

// userRepos accepts a user account and a set of repos. It returns a map from repository external ID
// to true/false indicating whether the given user has read access to each repo. If a repo ID is missing
// from the return map, the user does not have read access to that repo.
func (p *Provider) userRepos(ctx context.Context, userAccount *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (isAllowed map[string]bool, err error) {
	if userAccount == nil {
		return nil, nil
	}
	cachedUserRepos, err := p.getCachedUserRepos(ctx, userAccount, repos)
	if err != nil {
		return nil, err
	}
	if len(cachedUserRepos) == len(repos) {
		return cachedUserRepos, nil
	}

	missing := make(map[string]struct{})
	for r := range repos {
		if _, ok := cachedUserRepos[r.ExternalRepoSpec.ID]; !ok {
			missing[r.ExternalRepoSpec.ID] = struct{}{}
		}
	}

	uncachedUserRepos := make(map[string]bool)
	publicRepos := make(map[string]bool)
	for r := range missing {
		canAccess, isPublic, err := p.fetchUserRepo(ctx, userAccount, r)
		if err != nil {
			return nil, err
		}
		uncachedUserRepos[r] = canAccess
		publicRepos[r] = isPublic
	}

	if err := p.setCachedUserRepos(ctx, userAccount, uncachedUserRepos); err != nil {
		return nil, err
	}
	if err := p.setCachedPublicRepos(ctx, publicRepos); err != nil { // also cache whether repos are public
		return nil, err
	}
	for k, v := range uncachedUserRepos {
		cachedUserRepos[k] = v
	}
	return cachedUserRepos, nil
}

// setCachedUserRepos updates the cache with a map from GitHub repo ID to true/false indicating
// whether the user can access the repo. The GitHub repo ID is the GraphQL API ID ("repository node
// ID").
//
// Internally, it sets a separate cache key for each user and repo ID.
func (p *Provider) setCachedUserRepos(ctx context.Context, userAccount *extsvc.ExternalAccount, isAllowed map[string]bool) error {
	setArgs := make([][2]string, 0, len(isAllowed))
	for k, v := range isAllowed {
		rkey, err := json.Marshal(userRepoCacheKey{User: userAccount.AccountID, Repo: k})
		if err != nil {
			return err
		}
		rval, err := json.Marshal(userRepoCacheVal{Read: v, TTL: p.cacheTTL})
		if err != nil {
			return err
		}
		setArgs = append(setArgs, [2]string{string(rkey), string(rval)})
	}
	p.cache.SetMulti(setArgs...)
	return nil
}

// getCachedUserRepos accepts a user account and set of repos and returns a map from repo ID to
// true/false indicating whether the user can access the repo. The returned map may be incomplete
// (i.e., not every input repo may be represented in the key set) due to cache incompleteness.
func (p *Provider) getCachedUserRepos(ctx context.Context, userAccount *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[string]bool, error) {
	getArgs := make([]string, 0, len(repos))
	repoList := make([]string, 0, len(repos))
	for repo := range repos {
		rkey, err := json.Marshal(userRepoCacheKey{
			User: userAccount.AccountID,
			Repo: repo.ExternalRepoSpec.ID,
		})
		if err != nil {
			return nil, err
		}
		getArgs = append(getArgs, string(rkey))
		repoList = append(repoList, repo.ExternalRepoSpec.ID)
	}

	cacheVals := p.cache.GetMulti(getArgs...)
	if len(cacheVals) == 0 {
		return nil, nil
	}
	cachedIsAllowed := make(map[string]bool)
	for i, v := range cacheVals {
		if len(v) == 0 {
			continue
		}

		var val userRepoCacheVal
		if err := json.Unmarshal(v, &val); err != nil {
			return nil, err
		}
		cachedIsAllowed[repoList[i]] = val.Read
	}
	return cachedIsAllowed, nil
}

// fetchUserRepo fetches whether the given user can access the given repo from the GitHub API.
func (p *Provider) fetchUserRepo(ctx context.Context, userAccount *extsvc.ExternalAccount, repoID string) (canAccess bool, isPublic bool, err error) {
	_, tok, err := github.GetExternalAccountData(&userAccount.ExternalAccountData)
	if err != nil {
		return false, false, err
	}
	ghRepo, err := p.client.GetRepositoryByNodeID(ctx, tok.AccessToken, repoID)
	if err != nil {
		if err == github.ErrNotFound {
			return false, false, nil
		}
		return false, false, err
	}
	return true, !ghRepo.IsPrivate, nil
}

// FetchAccount implements the authz.Provider interface. It always returns nil, because the GitHub
// API doesn't currently provide a way to fetch user by external SSO account.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.ExternalAccount) (mine *extsvc.ExternalAccount, err error) {
	return nil, nil
}

func (p *Provider) ServiceID() string {
	return p.codeHost.ServiceID()
}

func (p *Provider) ServiceType() string {
	return p.codeHost.ServiceType()
}

func (p *Provider) Validate() (problems []string) {
	return nil
}
