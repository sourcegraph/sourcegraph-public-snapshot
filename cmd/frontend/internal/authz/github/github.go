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

type Provider struct {
	client   *github.Client
	codeHost *github.CodeHost
	cacheTTL time.Duration
	cache    pcache
}

type pcache interface {
	GetMulti(keys ...string) [][]byte
	SetMulti(keyvals ...[2]string)
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
	Delete(key string)
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

func (p *Provider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	return authz.GetCodeHostRepos(p.codeHost, repos)
}

type userRepoCacheVal struct {
	Read bool
	TTL  time.Duration
}

type publicRepoCacheVal struct {
	Public bool
	TTL    time.Duration
}

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

func (p *Provider) publicRepos(ctx context.Context, repos map[authz.Repo]struct{}) (map[string]bool, error) {
	cachedIsPublic, err := p.getCachedPublicRepos(ctx, repos)
	if err != nil {
		return nil, err
	}
	if len(cachedIsPublic) >= len(repos) {
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

func (p *Provider) setCachedPublicRepos(ctx context.Context, isPublic map[string]bool) error {
	setArgs := make([][2]string, 0, len(isPublic))
	for k, v := range isPublic {
		key := fmt.Sprintf("r:%s", k)
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

func (p *Provider) getCachedPublicRepos(ctx context.Context, repos map[authz.Repo]struct{}) (isPublic map[string]bool, err error) {
	if len(repos) == 0 {
		return nil, nil
	}
	isPublic = make(map[string]bool)
	repoList := make([]string, 0, len(repos))
	getArgs := make([]string, 0, len(repos))
	for r := range repos {
		getArgs = append(getArgs, fmt.Sprintf("r:%s", r))
		repoList = append(repoList, r.ExternalRepoSpec.ID)
	}
	vals := p.cache.GetMulti(getArgs...)
	if len(vals) != len(repos) {
		return nil, fmt.Errorf("number of cache items did not match number of keys")
	}

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

// fetchPublicRepos returns a map where the keys are GitHub repository node IDs and the values are booleans
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

func (p *Provider) userRepos(ctx context.Context, userAccount *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (isAllowed map[string]bool, err error) {
	if userAccount == nil {
		return nil, nil
	}
	cachedUserRepos, err := p.getCachedUserRepos(ctx, userAccount, repos)
	if err != nil {
		return nil, err
	}
	if len(cachedUserRepos) >= len(repos) {
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

func (p *Provider) setCachedUserRepos(ctx context.Context, userAccount *extsvc.ExternalAccount, isAllowed map[string]bool) error {
	setArgs := make([][2]string, 0, len(isAllowed))
	for k, v := range isAllowed {
		rkey, err := json.Marshal(struct {
			User string
			Repo string
		}{userAccount.AccountID, k})
		if err != nil {
			return err
		}
		rval, err := json.Marshal(userRepoCacheVal{
			Read: v,
			TTL:  p.cacheTTL,
		})
		if err != nil {
			return err
		}
		setArgs = append(setArgs, [2]string{string(rkey), string(rval)})
	}
	p.cache.SetMulti(setArgs...)
	return nil
}

func (p *Provider) getCachedUserRepos(ctx context.Context, userAccount *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[string]bool, error) {
	getArgs := make([]string, 0, len(repos))
	repoList := make([]string, 0, len(repos))
	for repo := range repos {
		rkey, err := json.Marshal(struct {
			User string
			Repo string
		}{userAccount.AccountID, repo.ExternalRepoSpec.ID})
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

// FetchAccount always returns nil, because the GitHub API doesn't currently provide a way to fetch user by external SSO account.
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
