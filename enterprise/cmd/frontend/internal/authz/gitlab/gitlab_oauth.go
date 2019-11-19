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
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"gopkg.in/inconshreveable/log15.v2"
)

var _ authz.Provider = (*OAuthAuthzProvider)(nil)

type OAuthAuthzProvider struct {
	clientProvider *gitlab.ClientProvider
	clientURL      *url.URL
	codeHost       *extsvc.CodeHost
	cache          cache
	cacheTTL       time.Duration
}

type OAuthAuthzProviderOp struct {
	// BaseURL is the URL of the GitLab instance.
	BaseURL *url.URL

	// CacheTTL is the TTL of cached permissions lists from the GitLab API.
	CacheTTL time.Duration

	// MockCache, if non-nil, replaces the default Redis-based cache with the supplied cache mock.
	// Should only be used in tests.
	MockCache cache
}

func newOAuthProvider(op OAuthAuthzProviderOp) *OAuthAuthzProvider {
	p := &OAuthAuthzProvider{
		clientProvider: gitlab.NewClientProvider(op.BaseURL, nil),
		clientURL:      op.BaseURL,
		codeHost:       extsvc.NewCodeHost(op.BaseURL, gitlab.ServiceType),
		cache:          op.MockCache,
		cacheTTL:       op.CacheTTL,
	}
	if p.cache == nil {
		p.cache = rcache.NewWithTTL(fmt.Sprintf("gitlabAuthz:%s", op.BaseURL.String()), int(math.Ceil(op.CacheTTL.Seconds())))
	}
	return p
}

func (p *OAuthAuthzProvider) Validate() (problems []string) {
	return nil
}

func (p *OAuthAuthzProvider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *OAuthAuthzProvider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *OAuthAuthzProvider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.ExternalAccount) (mine *extsvc.ExternalAccount, err error) {
	return nil, nil
}

func (p *OAuthAuthzProvider) RepoPerms(ctx context.Context, account *extsvc.ExternalAccount, repos []*types.Repo) (
	[]authz.RepoPerms, error,
) {
	accountID := "" // empty means public / unauthenticated to the code host
	if account != nil && account.ServiceID == p.codeHost.ServiceID && account.ServiceType == p.codeHost.ServiceType {
		accountID = account.AccountID
	}

	remaining := repos
	perms := make([]authz.RepoPerms, 0, len(remaining))

	// Populate perms using cached repository visibility information.
	for _, repo := range remaining {
		projID, err := strconv.Atoi(repo.ExternalRepo.ID)
		if err != nil {
			return nil, errors.Wrap(err, "GitLab repo external ID did not parse to int")
		}

		if vis, exists := cacheGetRepoVisibility(p.cache, projID, p.cacheTTL); exists {
			if v := vis.Visibility; v == gitlab.Public || (v == gitlab.Internal && accountID != "") {
				perms = append(perms, authz.RepoPerms{Repo: repo, Perms: authz.Read})
				continue
			}
		}

		// Populate perms using cached user-can-access-repository information.
		if accountID != "" {
			if userRepo, exists := cacheGetUserRepo(p.cache, accountID, projID, p.cacheTTL); exists {
				rp := authz.RepoPerms{Repo: repo}
				if userRepo.Read {
					rp.Perms = authz.Read
				}
				perms = append(perms, rp)
				continue
			}
		}

		// Populate perms for the remaining repos (nextRemaining) by fetching directly from the GitLab
		// API (and update the user repo-visibility and user-can-access-repo permissions, as well)
		var oauthToken string
		if account != nil {
			_, tok, err := gitlab.GetExternalAccountData(&account.ExternalAccountData)
			if err != nil {
				return nil, err
			}
			oauthToken = tok.AccessToken
		}

		isAccessible, vis, isContentAccessible, err := p.fetchProjVis(ctx, oauthToken, projID)
		if err != nil {
			log15.Error("Failed to fetch visibility for GitLab project", "projectID", projID, "gitlabHost", p.codeHost.BaseURL.String(), "error", err)
			continue
		}
		if isAccessible {
			// Set perms
			perms = append(perms, authz.RepoPerms{Repo: repo, Perms: authz.Read})

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

// fetchProjVis fetches a repository's visibility with usr's credentials. It returns:
// - whether the project is accessible to the user,
// - the visibility if the repo is accessible (otherwise this is empty),
// - whether the repository contents are accessible to usr, and
// - any error encountered in fetching (not including an error due to the repository not being visible);
//   if the error is non-nil, all other return values should be disregraded
func (p *OAuthAuthzProvider) fetchProjVis(ctx context.Context, oauthToken string, projID int) (
	isAccessible bool, vis gitlab.Visibility, isContentAccessible bool, err error,
) {
	proj, err := p.clientProvider.GetOAuthClient(oauthToken).GetProject(ctx, gitlab.GetProjectOp{
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

	if _, err := p.clientProvider.GetOAuthClient(oauthToken).ListTree(ctx, gitlab.ListTreeOp{
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
