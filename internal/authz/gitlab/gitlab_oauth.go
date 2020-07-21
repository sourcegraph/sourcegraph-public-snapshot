// Package gitlab contains an authorization provider for GitLab that uses GitLab OAuth
// authenetication.
package gitlab

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

var _ authz.Provider = (*OAuthProvider)(nil)

type OAuthProvider struct {
	// The token is the access token used for syncing repositories from the code host,
	// but it may or may not be a sudo-scoped.
	token string

	urn               string
	clientProvider    *gitlab.ClientProvider
	clientURL         *url.URL
	codeHost          *extsvc.CodeHost
	cache             cache
	cacheTTL          time.Duration
	minBatchThreshold int
	maxBatchRequests  int
}

type OAuthProviderOp struct {
	// The unique resource identifier of the external service where the provider is defined.
	URN string

	// BaseURL is the URL of the GitLab instance.
	BaseURL *url.URL

	// Token is an access token with api scope, it may or may not have sudo scope.
	//
	// 🚨 SECURITY: This value contains secret information that must not be shown to non-site-admins.
	Token string

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

func newOAuthProvider(op OAuthProviderOp, cli httpcli.Doer) *OAuthProvider {
	p := &OAuthProvider{
		token: op.Token,

		urn:               op.URN,
		clientProvider:    gitlab.NewClientProvider(op.BaseURL, cli),
		clientURL:         op.BaseURL,
		codeHost:          extsvc.NewCodeHost(op.BaseURL, extsvc.TypeGitLab),
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

func (p *OAuthProvider) Validate() (problems []string) {
	return nil
}

func (p *OAuthProvider) URN() string {
	return p.urn
}

func (p *OAuthProvider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *OAuthProvider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *OAuthProvider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account) (mine *extsvc.Account, err error) {
	return nil, nil
}

// fetchProjVis fetches a repository's visibility with usr's credentials. It returns:
// - whether the project is accessible to the user,
// - the visibility if the repo is accessible (otherwise this is empty),
// - whether the repository contents are accessible to usr, and
// - any error encountered in fetching (not including an error due to the repository not being visible);
//   if the error is non-nil, all other return values should be disregraded
func (p *OAuthProvider) fetchProjVis(ctx context.Context, oauthToken string, projID int) (
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
func (p *OAuthProvider) fetchProjVisBatch(ctx context.Context, oauthToken string, reposByProjID map[int]*types.Repo, op fetchProjectVisibilityBatchOp) (
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
