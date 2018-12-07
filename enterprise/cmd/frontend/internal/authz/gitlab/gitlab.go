// Package gitlab contains an authorization provider for GitLab.
package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type pcache interface {
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
	Delete(key string)
}

// GitLabAuthzProvider is an implementation of AuthzProvider that provides repository permissions as
// determined from a GitLab instance API. For documentation of specific fields, see the docstrings
// of GitLabAuthzProviderOp.
type GitLabAuthzProvider struct {
	client            *gitlab.Client
	clientURL         *url.URL
	codeHost          *gitlab.CodeHost
	gitlabProvider    string
	authnConfigID     auth.ProviderConfigID
	useNativeUsername bool
	cache             pcache
	cacheTTL          time.Duration
}

var _ authz.Provider = ((*GitLabAuthzProvider)(nil))

type cacheVal struct {
	// ProjIDs is the set of project IDs to which a GitLab user has access.
	ProjIDs map[int]struct{} `json:"repos"`

	// TTL is the ttl of the cache entry. This must be checked for equality in case the TTL has
	// changed (and the cache entry should therefore be invalidated).
	TTL time.Duration `json:"ttl"`
}

type GitLabAuthzProviderOp struct {
	// BaseURL is the URL of the GitLab instance.
	BaseURL *url.URL

	// AuthnConfigID identifies the authn provider to use to lookup users on the GitLab instance.
	// This should be the authn provider that's used to sign into the GitLab instance.
	AuthnConfigID auth.ProviderConfigID

	// GitLabProvider is the id of the authn provider to GitLab. It will be used in the
	// `users?extern_uid=$uid&provider=$provider` API query.
	GitLabProvider string

	// SudoToken is an access token with sudo *and* api scope.
	//
	// 🚨 SECURITY: This value contains secret information that must not be shown to non-site-admins.
	SudoToken string

	// CacheTTL is the TTL of cached permissions lists from the GitLab API.
	CacheTTL time.Duration

	// UseNativeUsername, if true, maps Sourcegraph users to GitLab users using username equivalency
	// instead of the authn provider user ID. This is *very* insecure (Sourcegraph usernames can be
	// changed at the user's will) and should only be used in development environments.
	UseNativeUsername bool

	// MockCache, if non-nil, replaces the default Redis-based cache with the supplied cache mock.
	// Should only be used in tests.
	MockCache pcache
}

func NewProvider(op GitLabAuthzProviderOp) *GitLabAuthzProvider {
	p := &GitLabAuthzProvider{
		client:            gitlab.NewClient(op.BaseURL, op.SudoToken, "", nil),
		clientURL:         op.BaseURL,
		codeHost:          gitlab.NewCodeHost(op.BaseURL),
		cache:             op.MockCache,
		authnConfigID:     op.AuthnConfigID,
		gitlabProvider:    op.GitLabProvider,
		cacheTTL:          op.CacheTTL,
		useNativeUsername: op.UseNativeUsername,
	}
	if p.cache == nil {
		p.cache = rcache.NewWithTTL(fmt.Sprintf("gitlabAuthz:%s", op.BaseURL.String()), int(math.Ceil(op.CacheTTL.Seconds())))
	}
	return p
}

func (p *GitLabAuthzProvider) Validate() (problems []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, _, err := p.client.ListProjects(ctx, "projects?sudo=1"); err != nil {
		if err == ctx.Err() {
			problems = append(problems, fmt.Sprintf("GitLab API did not respond within 5s (%s)", err.Error()))
		} else if !gitlab.IsNotFound(err) {
			problems = append(problems, "access token did not have sufficient privileges, requires scopes \"sudo\" and \"api\"")
		}
	}
	return problems
}

func (p *GitLabAuthzProvider) ServiceID() string {
	return p.codeHost.ServiceID()
}

func (p *GitLabAuthzProvider) ServiceType() string {
	return p.codeHost.ServiceType()
}

func (p *GitLabAuthzProvider) RepoPerms(ctx context.Context, account *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (map[api.RepoName]map[authz.Perm]bool, error) {
	accountID := "" // empty means public / unauthenticated to the code host
	if account != nil && account.ServiceID == p.codeHost.ServiceID() && account.ServiceType == p.codeHost.ServiceType() {
		accountID = account.AccountID
	}

	myRepos, _ := p.Repos(ctx, repos)
	var accessibleRepos map[int]struct{}
	if r, exists := p.getCachedAccessList(accountID); exists {
		accessibleRepos = r
	} else {
		var err error
		accessibleRepos, err = p.fetchUserAccessList(ctx, accountID)
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

func (p *GitLabAuthzProvider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	return authz.GetCodeHostRepos(p.codeHost, repos)
}

func (p *GitLabAuthzProvider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.ExternalAccount) (mine *extsvc.ExternalAccount, err error) {
	if user == nil {
		return nil, nil
	}

	var glUser *gitlab.User
	if p.useNativeUsername {
		glUser, err = p.fetchAccountByUsername(ctx, user.Username)
	} else {
		// resolve the GitLab account using the authn provider (specified by p.AuthnConfigID)
		authnProvider := getProviderByConfigID(p.authnConfigID)
		if authnProvider == nil {
			return nil, nil
		}
		var authnAcct *extsvc.ExternalAccount
		for _, acct := range current {
			if acct.ServiceID == authnProvider.CachedInfo().ServiceID && acct.ServiceType == authnProvider.ConfigID().Type {
				authnAcct = acct
				break
			}
		}
		if authnAcct == nil {
			return nil, nil
		}
		glUser, err = p.fetchAccountByExternalUID(ctx, authnAcct.AccountID)
	}
	if err != nil {
		return nil, err
	}
	if glUser == nil {
		return nil, nil
	}

	jsonGLUser, err := json.Marshal(glUser)
	if err != nil {
		return nil, err
	}
	accountData := json.RawMessage(jsonGLUser)
	glExternalAccount := extsvc.ExternalAccount{
		UserID: user.ID,
		ExternalAccountSpec: extsvc.ExternalAccountSpec{
			ServiceType: p.codeHost.ServiceType(),
			ServiceID:   p.codeHost.ServiceID(),
			AccountID:   strconv.Itoa(int(glUser.ID)),
		},
		ExternalAccountData: extsvc.ExternalAccountData{
			AccountData: &accountData,
		},
	}
	return &glExternalAccount, nil
}

func (p *GitLabAuthzProvider) fetchAccountByExternalUID(ctx context.Context, uid string) (*gitlab.User, error) {
	q := make(url.Values)
	q.Add("extern_uid", uid)
	q.Add("provider", p.gitlabProvider)
	q.Add("per_page", "2")
	glUsers, _, err := p.client.ListUsers(ctx, "users?"+q.Encode())
	if err != nil {
		return nil, err
	}
	if len(glUsers) >= 2 {
		return nil, fmt.Errorf("failed to determine unique GitLab user for query %q", q.Encode())
	}
	if len(glUsers) == 0 {
		return nil, nil
	}
	return glUsers[0], nil
}

func (p *GitLabAuthzProvider) fetchAccountByUsername(ctx context.Context, username string) (*gitlab.User, error) {
	q := make(url.Values)
	q.Add("username", username)
	q.Add("per_page", "2")
	glUsers, _, err := p.client.ListUsers(ctx, "users?"+q.Encode())
	if err != nil {
		return nil, err
	}
	if len(glUsers) >= 2 {
		return nil, fmt.Errorf("failed to determine unique GitLab user for query %q", q.Encode())
	}
	if len(glUsers) == 0 {
		return nil, nil
	}
	return glUsers[0], nil
}

// getCachedAccessList returns the list of repositories accessible to a user from the cache and
// whether the cache entry exists.
func (p *GitLabAuthzProvider) getCachedAccessList(accountID string) (map[int]struct{}, bool) {
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
func (p *GitLabAuthzProvider) fetchUserAccessList(ctx context.Context, glUserID string) (map[int]struct{}, error) {
	q := make(url.Values)
	if glUserID != "" {
		q.Add("sudo", glUserID)
	} else {
		q.Add("visibility", "public")
	}
	q.Add("per_page", "100")

	projIDs := make(map[int]struct{})
	var iters = 0
	var pageURL = "projects?" + q.Encode()
	for {
		if iters >= 100 && iters%100 == 0 {
			log15.Warn("Excessively many GitLab API requests to fetch complete user authz list", "iters", iters, "gitlabUserID", glUserID, "host", p.clientURL.String())
		}

		projs, nextPageURL, err := p.client.ListProjects(ctx, pageURL)
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

var getProviderByConfigID = auth.GetProviderByConfigID
