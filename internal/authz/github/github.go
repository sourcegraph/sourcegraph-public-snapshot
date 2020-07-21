package github

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

// Provider implements authz.Provider for GitHub repository permissions.
type Provider struct {
	urn      string
	client   client
	codeHost *extsvc.CodeHost
	cacheTTL time.Duration
	cache    cache
}

func NewProvider(urn string, githubURL *url.URL, baseToken string, client *github.Client, cacheTTL time.Duration, mockCache cache) *Provider {
	if client == nil {
		apiURL, _ := github.APIRoot(githubURL)
		client = github.NewClient(apiURL, baseToken, nil)
	}

	p := &Provider{
		urn:      urn,
		codeHost: extsvc.NewCodeHost(githubURL, extsvc.TypeGitHub),
		client:   &ClientAdapter{Client: client},
		cache:    mockCache,
		cacheTTL: cacheTTL,
	}

	// Note: this will use the same underlying Redis instance and key namespace for every instance
	// of Provider. This is by design, so that different instances, even in different processes,
	// will share cache entries.
	if p.cache == nil {
		p.cache = rcache.NewWithTTL(fmt.Sprintf("githubAuthz:%s", githubURL.String()), int(math.Ceil(cacheTTL.Seconds())))
	}
	return p
}

var _ authz.Provider = (*Provider)(nil)

// fetchAndSetPublicRepos accepts a set of repositories and returns a map from repository external
// ID (the GitHub repository GraphQL ID) to true/false indicating whether the repository is public
// or private. It consults and updates the cache. As a side effect, it caches the publicness of the
// repos.
func (p *Provider) fetchAndSetPublicRepos(ctx context.Context, repos []*types.Repo) (map[string]bool, error) {
	isPublic, err := p.fetchPublicRepos(ctx, repos)
	if err != nil {
		return nil, err
	}
	if err := p.setCachedPublicRepos(ctx, isPublic); err != nil {
		return nil, err
	}
	return isPublic, nil
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
func (p *Provider) getCachedPublicRepos(ctx context.Context, repos []*types.Repo) (isPublic map[string]bool, err error) {
	if len(repos) == 0 {
		return nil, nil
	}
	isPublic = make(map[string]bool)
	repoList := make([]string, 0, len(repos))
	getArgs := make([]string, 0, len(repos))
	for _, r := range repos {
		getArgs = append(getArgs, publicRepoCacheKey(r.ExternalRepo.ID))
		repoList = append(repoList, r.ExternalRepo.ID)
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
		if p.cacheTTL < val.TTL {
			// if the cache TTL is now less than the cache entry TTL, invalidate that entry
			continue
		}
		isPublic[repoList[i]] = val.Public
	}
	return isPublic, nil
}

// fetchPublicRepos returns a map from GitHub repository ID (the GraphQL repo node ID) to true/false
// indicating whether a repository is public (true) or private (false).
func (p *Provider) fetchPublicRepos(ctx context.Context, repos []*types.Repo) (map[string]bool, error) {
	isPublic := make(map[string]bool)
	for _, repo := range repos {
		ghRepo, err := p.client.GetRepositoryByNodeID(ctx, repo.ExternalRepo.ID)
		if err == github.ErrNotFound {
			// Note: we could set `isPublic[repo.ExternalRepoSpec.ID] = false` here, but
			// purposefully don't cache if a repo is private in case it later becomes public.
			continue
		}
		if err != nil {
			return nil, err
		}
		isPublic[repo.ExternalRepo.ID] = !ghRepo.IsPrivate
	}
	return isPublic, nil
}

// fetchAndSetUserRepos accepts a user account and a set of repos. It returns a map from repository
// external ID to true/false indicating whether the given user has read access to each repo. If a
// repo ID is missing from the return map, the user does not have read access to that repo. As a
// side effect, it caches the fetched repos (whether the given user has access to each and whether
// each is public).
func (p *Provider) fetchAndSetUserRepos(ctx context.Context, userAccount *extsvc.Account, repos []*types.Repo) (isAllowed map[string]bool, err error) {
	if userAccount == nil {
		return nil, nil
	}

	repoIDs := make([]string, len(repos))
	i := 0
	for _, repo := range repos {
		repoIDs[i] = repo.ExternalRepo.ID
		i++
	}

	canAccess, isPublic, err := p.fetchUserRepos(ctx, userAccount, repoIDs)
	if err != nil {
		return nil, err
	}
	userRepos := make(map[string]bool)
	publicRepos := make(map[string]bool)
	for _, r := range repos {
		userRepos[r.ExternalRepo.ID] = canAccess[r.ExternalRepo.ID]
		publicRepos[r.ExternalRepo.ID] = isPublic[r.ExternalRepo.ID]
	}

	if err := p.setCachedUserRepos(ctx, userAccount, userRepos); err != nil {
		return nil, err
	}
	if err := p.setCachedPublicRepos(ctx, publicRepos); err != nil { // also cache whether repos are public
		return nil, err
	}
	return userRepos, nil
}

// setCachedUserRepos updates the cache with a map from GitHub repo ID to true/false indicating
// whether the user can access the repo. The GitHub repo ID is the GraphQL API ID ("repository node
// ID").
//
// Internally, it sets a separate cache key for each user and repo ID.
func (p *Provider) setCachedUserRepos(ctx context.Context, userAccount *extsvc.Account, isAllowed map[string]bool) error {
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
func (p *Provider) getCachedUserRepos(ctx context.Context, userAccount *extsvc.Account, repos []*types.Repo) (cachedUserRepos map[string]bool, err error) {
	if userAccount == nil {
		return nil, nil
	}

	getArgs := make([]string, 0, len(repos))
	repoList := make([]string, 0, len(repos))
	for _, repo := range repos {
		rkey, err := json.Marshal(userRepoCacheKey{
			User: userAccount.AccountID,
			Repo: repo.ExternalRepo.ID,
		})
		if err != nil {
			return nil, err
		}
		getArgs = append(getArgs, string(rkey))
		repoList = append(repoList, repo.ExternalRepo.ID)
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
		if p.cacheTTL < val.TTL {
			// if the cache TTL is now less than the cache entry TTL, invalidate that entry
			continue
		}
		cachedIsAllowed[repoList[i]] = val.Read
	}
	return cachedIsAllowed, nil
}

func (p *Provider) fetchUserRepos(ctx context.Context, userAccount *extsvc.Account, repoIDs []string) (canAccess map[string]bool, isPublic map[string]bool, err error) {
	_, tok, err := github.GetExternalAccountData(&userAccount.AccountData)
	if err != nil {
		return nil, nil, err
	}
	client := p.client.WithToken(tok.AccessToken)

	// Batch fetch repos from API
	ghRepos := make(map[string]*github.Repository)
	for i := 0; i < len(repoIDs); i += github.MaxNodeIDs {
		j := i + github.MaxNodeIDs
		if j > len(repoIDs) {
			j = len(repoIDs)
		}
		ghReposBatch, err := client.GetRepositoriesByNodeIDFromAPI(ctx, repoIDs[i:j])
		if err != nil {
			return nil, nil, err
		}
		for k, v := range ghReposBatch {
			ghRepos[k] = v
		}
	}

	isPublic = make(map[string]bool)
	for _, r := range ghRepos {
		isPublic[r.ID] = !r.IsPrivate
	}
	canAccess = make(map[string]bool)
	for _, rid := range repoIDs {
		_, exists := ghRepos[rid]
		canAccess[rid] = exists
	}

	return canAccess, isPublic, nil
}

// FetchAccount implements the authz.Provider interface. It always returns nil, because the GitHub
// API doesn't currently provide a way to fetch user by external SSO account.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account) (mine *extsvc.Account, err error) {
	return nil, nil
}

func (p *Provider) URN() string {
	return p.urn
}

func (p *Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *Provider) Validate() (problems []string) {
	return nil
}

// FetchUserPerms returns a list of repository IDs (on code host) that the given account
// has read access on the code host. The repository ID has the same value as it would be
// used as api.ExternalRepoSpec.ID. The returned list only includes private repository IDs.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://developer.github.com/v3/repos/#list-repositories-for-the-authenticated-user
func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account) ([]extsvc.RepoID, error) {
	if account == nil {
		return nil, errors.New("no account provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, account) {
		return nil, fmt.Errorf("not a code host of the account: want %q but have %q",
			account.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	_, tok, err := github.GetExternalAccountData(&account.AccountData)
	if err != nil {
		return nil, errors.Wrap(err, "get external account data")
	} else if tok == nil {
		return nil, errors.New("no token found in the external account data")
	}

	// 🚨 SECURITY: Use user token is required to only list repositories the user has access to.
	client := p.client.WithToken(tok.AccessToken)

	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	repoIDs := make([]extsvc.RepoID, 0, 100)
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var repos []*github.Repository
		repos, hasNextPage, _, err = client.ListAffiliatedRepositories(ctx, github.VisibilityPrivate, page)
		if err != nil {
			return repoIDs, err
		}

		for _, r := range repos {
			repoIDs = append(repoIDs, extsvc.RepoID(r.ID))
		}
	}

	return repoIDs, nil
}

// FetchRepoPerms returns a list of user IDs (on code host) who have read access to
// the given project on the code host. The user ID has the same value as it would
// be used as extsvc.Account.AccountID. The returned list includes both direct access
// and inherited from the organization membership.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://developer.github.com/v4/object/repositorycollaboratorconnection/
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository) ([]extsvc.AccountID, error) {
	if repo == nil {
		return nil, errors.New("no repository provided")
	} else if !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternalRepoSpec) {
		return nil, fmt.Errorf("not a code host of the repository: want %q but have %q",
			repo.ServiceID, p.codeHost.ServiceID)
	}

	// NOTE: We do not store port or scheme in our URI, so stripping the hostname alone is enough.
	nameWithOwner := strings.TrimPrefix(repo.URI, p.codeHost.BaseURL.Hostname())
	nameWithOwner = strings.TrimPrefix(nameWithOwner, "/")

	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return nil, errors.Wrap(err, "split nameWithOwner")
	}

	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	userIDs := make([]extsvc.AccountID, 0, 100)
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var err error
		var users []*github.Collaborator
		users, hasNextPage, err = p.client.ListRepositoryCollaborators(ctx, owner, name, page)
		if err != nil {
			return userIDs, err
		}

		for _, u := range users {
			userIDs = append(userIDs, extsvc.AccountID(strconv.FormatInt(u.DatabaseID, 10)))
		}
	}

	return userIDs, nil
}
