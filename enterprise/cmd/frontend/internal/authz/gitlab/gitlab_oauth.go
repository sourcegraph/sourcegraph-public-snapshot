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
	clientProvider    *gitlab.ClientProvider
	clientURL         *url.URL
	codeHost          *extsvc.CodeHost
	cache             cache
	cacheTTL          time.Duration
	minBatchThreshold int
	maxBatchRequests  int
}

type OAuthAuthzProviderOp struct {
	// BaseURL is the URL of the GitLab instance.
	BaseURL *url.URL

	// CacheTTL is the TTL of cached permissions lists from the GitLab API.
	CacheTTL time.Duration

	// MockCache, if non-nil, replaces the default Redis-based cache with the supplied cache mock.
	// Should only be used in tests.
	MockCache cache

	// MinBatchThreshold is the number of repositories at which we start trying to batch fetch
	// GitLab project visibility. This should be in the neighborhood of maxBatchRequests, because
	// batch-fetching means we fetch *all* projects visible to the given user (not just the ones
	// requested in RepoPerms)
	MinBatchThreshold int

	// MaxBatchRequests is the maximum number of batch requests we make for GitLab project
	// visibility. We limit this in case the user has access to many more projects than are being
	// requested in RepoPerms.
	MaxBatchRequests int
}

func newOAuthProvider(op OAuthAuthzProviderOp) *OAuthAuthzProvider {
	p := &OAuthAuthzProvider{
		clientProvider:    gitlab.NewClientProvider(op.BaseURL, nil),
		clientURL:         op.BaseURL,
		codeHost:          extsvc.NewCodeHost(op.BaseURL, gitlab.ServiceType),
		cache:             op.MockCache,
		cacheTTL:          op.CacheTTL,
		minBatchThreshold: op.MinBatchThreshold,
		maxBatchRequests:  op.MaxBatchRequests,
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

	reposByProjID := make(map[int]*types.Repo, len(repos))
	for _, repo := range repos {
		projID, err := strconv.Atoi(repo.ExternalRepo.ID)
		if err != nil {
			return nil, errors.Wrap(err, "GitLab repo external ID did not parse to int")
		}
		reposByProjID[projID] = repo
	}

	// remaining tracks which repositories permissions remain to be computed for, keyed by project ID
	remaining := make(map[int]*types.Repo, len(repos))
	perms := make([]authz.RepoPerms, 0, len(repos))

	// Populate perms using cached repository visibility information.
	for projID, repo := range reposByProjID {
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

		remaining[projID] = repo
	}

	var oauthToken string
	if account != nil {
		_, tok, err := gitlab.GetExternalAccountData(&account.ExternalAccountData)
		if err != nil {
			return nil, err
		}
		oauthToken = tok.AccessToken
	}

	// Best-effort fetch visibility in batch if we have more than X remaining repositories to check
	// and user is authenticated.
	//
	// This is an optimization. If we have too many repositories (GitLab calls them "projects") to
	// fetch from the GitLab API, we try to batch-fetch all projects whose visibility is `internal`
	// or `public` (because we can batch-fetch 100 repositories at a time this way). This is not
	// guaranteed to be strictly better than fetching them individually, because if we batch-fest,
	// we must batch-fest *all* repositories, not just the ones in `repos`.
	//
	// We cannot determine the permissions of projects with visibility `private` this way, because a
	// project may be visible to a GitLab user, but its contents inaccessible (which means we have
	// to issue individual API requests to request repository contents to verify permissions).
	if len(remaining) >= p.minBatchThreshold && oauthToken != "" {
		nextRemaining := make(map[int]*types.Repo, len(remaining))
		visibility, err := p.fetchProjVisBatch(ctx, oauthToken, remaining, fetchProjectVisibilityBatchOp{maxRequests: p.maxBatchRequests})
		if err != nil {
			log15.Error("Error encountered fetching project visibility from GitLab", "err", err)
		}
		for projID, repo := range remaining {
			vis, ok := visibility[projID]
			if !(ok && (vis == visibilityPublic || vis == visibilityInternal)) {
				nextRemaining[projID] = repo
				continue
			}

			// Set perms (visibility is public or internal)
			perms = append(perms, authz.RepoPerms{Repo: repo, Perms: authz.Read})
		}
		remaining = nextRemaining
	}

	// Fetch individually
	for projID, repo := range remaining {
		// Populate perms for the remaining repos (`remaining`) by fetching directly from the GitLab
		// API (and update the user repo-visibility and user-can-access-repo permissions, as well)

		isAccessible, vis, isContentAccessible, err := p.fetchProjVis(ctx, oauthToken, projID)
		if err != nil {
			log15.Error("Failed to fetch visibility for GitLab project", "repoName", repo.Name, "projectID", projID, "gitlabHost", p.codeHost.BaseURL.String(), "error", err)
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

type fetchProjectVisibilityBatchOp struct {
	// maxRequests is the maximum number of requests to issue before returning
	maxRequests int
}

type visibilityLevel int

const (
	visibilityUnknown = iota
	visibilityPrivate
	visibilityInternal
	visibilityPublic
)

// batchProjVisSize is the number of projects to request visibility for in each batch request issued
// by fetchProjVisBatch
const batchProjVisSize = 100

// fetchProjVisBatch returns the list of repositories best-effort sorted into groups. The visiblity
// results are valid even if err is non-nil.
func (p *OAuthAuthzProvider) fetchProjVisBatch(ctx context.Context, oauthToken string, reposByProjID map[int]*types.Repo, op fetchProjectVisibilityBatchOp) (
	projIDVisibility map[int]visibilityLevel, err error,
) {
	projIDVisibility = make(map[int]visibilityLevel, len(reposByProjID))
	for projID := range reposByProjID {
		projIDVisibility[projID] = visibilityUnknown
	}

	matchCount := 0
	projPageURL := fmt.Sprintf("projects?per_page=%d", batchProjVisSize)
	for i := 0; i < op.maxRequests; i++ {
		if matchCount >= len(reposByProjID) {
			break
		}
		projs, next, err := p.clientProvider.GetOAuthClient(oauthToken).ListProjects(ctx, projPageURL)
		if err != nil {
			return projIDVisibility, err
		}

		for _, proj := range projs {
			if err := cacheSetRepoVisibility(p.cache, proj.ID, repoVisibilityCacheVal{Visibility: proj.Visibility, TTL: p.cacheTTL}); err != nil {
				log15.Error("could not set cached repo visibility from batch fetch", "projPath", proj.PathWithNamespace, "err", err)
			}

			projVis, ok := projIDVisibility[proj.ID]
			if !ok {
				continue
			}
			if projVis != visibilityUnknown {
				continue
			}
			switch proj.Visibility {
			case gitlab.Public:
				projIDVisibility[proj.ID] = visibilityPublic
			case gitlab.Internal:
				projIDVisibility[proj.ID] = visibilityInternal
			case gitlab.Private:
				projIDVisibility[proj.ID] = visibilityPrivate
			}
			matchCount++
		}

		if next == nil {
			break
		}
		projPageURL = *next
	}

	return projIDVisibility, nil
}
