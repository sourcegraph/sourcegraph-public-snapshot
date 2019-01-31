// Package gitlab contains an authorization provider for GitLab that uses GitLab OAuth
// authenetication.
package gitlab

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
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

func (p *GitLabOAuthAuthzProvider) Repos(ctx context.Context, repos map[authz.Repo]struct{}) (mine map[authz.Repo]struct{}, others map[authz.Repo]struct{}) {
	return authz.GetCodeHostRepos(p.codeHost, repos)
}

func (p *GitLabOAuthAuthzProvider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.ExternalAccount) (mine *extsvc.ExternalAccount, err error) {
	return nil, nil
}

func (p *GitLabOAuthAuthzProvider) RepoPerms(ctx context.Context, account *extsvc.ExternalAccount, repos map[authz.Repo]struct{}) (
	map[api.RepoName]map[authz.Perm]bool, error,
) {
	accountID := "" // empty means public / unauthenticated to the code host
	if account != nil && account.ServiceID == p.codeHost.ServiceID() && account.ServiceType == p.codeHost.ServiceType() {
		accountID = account.AccountID
	}

	perms := map[api.RepoName]map[authz.Perm]bool{}

	remaining, _ := p.Repos(ctx, repos)
	nextRemaining := map[authz.Repo]struct{}{}

	// Check cached visibility
	for repo := range remaining {
		projID, err := strconv.Atoi(repo.ExternalRepoSpec.ID)
		if err != nil {
			return nil, errors.Wrap(err, "GitLab repo external ID did not parse to int")
		}
		vis, exists := cacheGetRepoVisibility(p.cache, projID, p.cacheTTL)
		if !exists {
			nextRemaining[repo] = struct{}{}
			continue
		}
		switch v := vis.Visibility; {
		case v == gitlab.Public:
			fallthrough
		case v == gitlab.Internal && accountID != "":
			perms[repo.RepoName] = map[authz.Perm]bool{authz.Read: true}
			continue
		}
		nextRemaining[repo] = struct{}{}
	}

	if len(nextRemaining) == 0 { // shortcut
		return perms, nil
	}

	// Check cached user repos
	if accountID != "" {
		remaining, nextRemaining = nextRemaining, map[authz.Repo]struct{}{}
		for repo := range remaining {
			projID, err := strconv.Atoi(repo.ExternalRepoSpec.ID)
			if err != nil {
				return nil, errors.Wrap(err, "GitLab repo external ID did not parse to int")
			}
			userRepo, exists := cacheGetUserRepo(p.cache, accountID, projID, p.cacheTTL)
			if !exists {
				nextRemaining[repo] = struct{}{}
				continue
			}
			perms[repo.RepoName] = map[authz.Perm]bool{authz.Read: userRepo.Read}
		}

		if len(nextRemaining) == 0 { // shortcut
			return perms, nil
		}
	}

	// Fetch and update cache
	var accessToken string
	if account != nil {
		_, tok, err := gitlab.GetExternalAccountData(&account.ExternalAccountData)
		if err != nil {
			return nil, err
		}
		accessToken = tok.AccessToken
	}
	for repo := range remaining {
		projID, err := strconv.Atoi(repo.ExternalRepoSpec.ID)
		if err != nil {
			return nil, errors.Wrap(err, "GitLab repo external ID did not parse to int")
		}
		isAccessible, vis, isContentAccessible, err := p.fetchProjVis(ctx, accessToken, projID)
		if err != nil {
			log15.Error("Failed to fetch visibility for GitLab project", "projectID", projID, "gitlabHost", p.codeHost.BaseURL().String(), "error", err)
			continue
		}
		if isAccessible {
			// Set perms
			perms[repo.RepoName] = map[authz.Perm]bool{authz.Read: true}

			// Update visibility cache
			err := cacheSetRepoVisibility(p.cache, projID, repoVisibilityCacheVal{Visibility: vis, TTL: p.cacheTTL})
			if err != nil {
				return nil, errors.Wrap(err, "could not set cached repo visibility")
			}

			// Update userRepo cache if the visibility is private
			if vis == gitlab.Private {
				err := cacheSetUserRepo(p.cache, accountID, projID, userRepoCacheVal{Read: isContentAccessible, TTL: p.cacheTTL})
				if err != nil {
					return nil, errors.Wrap(err, "could not set cached user repo")
				}
			}
		} else if accountID != "" {
			// A repo is private if it is not accessible to an authenticated user
			err := cacheSetRepoVisibility(p.cache, projID, repoVisibilityCacheVal{Visibility: gitlab.Private, TTL: p.cacheTTL})
			if err != nil {
				return nil, errors.Wrap(err, "could not set cached repo visibility")
			}
			err = cacheSetUserRepo(p.cache, accountID, projID, userRepoCacheVal{Read: false, TTL: p.cacheTTL})
			if err != nil {
				return nil, errors.Wrap(err, "could not set cached user repo")
			}
		}
	}
	return perms, nil
}

// fetchRepoVisibility fetches a repository's visibility with usr's credentials. It returns:
// - whether the project is accessible to the user,
// - the visibility if the repo is accessible (otherwise this is empty),
// - whether the repository contents are accessible to usr, and
// - any error encountered in fetching (not including an error due to the repository not being visible);
//   if the error is non-nil, all other return values should be disregraded
func (p *GitLabOAuthAuthzProvider) fetchProjVis(ctx context.Context, accessToken string, projID int) (
	isAccessible bool, vis gitlab.Visibility, isContentAccessible bool, err error,
) {
	proj, err := p.clientProvider.GetOAuthClient(accessToken).GetProject(ctx, gitlab.GetProjectOp{
		ID:       projID,
		CommonOp: gitlab.CommonOp{NoCache: true},
	})
	if err != nil {
		if errCode := gitlab.HTTPErrorCode(err); errCode == http.StatusNotFound {
			return false, "", false, nil
		}
		return false, "", false, err
	}

	// If we get here, the project is accessible to the user (user has at least "Guest" permissions
	// on the project).

	if proj.Visibility == gitlab.Public || proj.Visibility == gitlab.Internal {
		// All authenticated users can read the contents of all internal/public projects
		// (https://docs.gitlab.com/ee/user/permissions.html).
		return true, proj.Visibility, true, nil
	}

	// If project visibility is private and its accessible to user, we still need to check if the user
	// can read the repository contents (i.e., does not merely have "Guest" permissions).

	if _, err := p.clientProvider.GetOAuthClient(accessToken).ListTree(ctx, gitlab.ListTreeOp{
		ProjID:   projID,
		CommonOp: gitlab.CommonOp{NoCache: true},
	}); err != nil {
		if errCode := gitlab.HTTPErrorCode(err); errCode == http.StatusNotFound {
			return true, proj.Visibility, false, nil
		}
		return false, "", false, err
	}
	return true, proj.Visibility, true, nil
}
