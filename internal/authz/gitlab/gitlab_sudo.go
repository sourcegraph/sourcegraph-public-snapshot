// Package gitlab contains an authorization provider for GitLab.
package gitlab

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

type pcache interface {
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
	Delete(key string)
}

// SudoProvider is an implementation of AuthzProvider that provides repository permissions as
// determined from a GitLab instance API. For documentation of specific fields, see the docstrings
// of SudoProviderOp.
type SudoProvider struct {
	// sudoToken is the sudo-scoped access token. This is different from the Sudo parameter, which
	// is set per client and defines which user to impersonate.
	sudoToken string

	urn               string
	clientProvider    *gitlab.ClientProvider
	clientURL         *url.URL
	codeHost          *extsvc.CodeHost
	gitlabProvider    string
	authnConfigID     providers.ConfigID
	useNativeUsername bool
	cache             pcache
	cacheTTL          time.Duration
}

var _ authz.Provider = (*SudoProvider)(nil)

type SudoProviderOp struct {
	// The unique resource identifier of the external service where the provider is defined.
	URN string

	// BaseURL is the URL of the GitLab instance.
	BaseURL *url.URL

	// AuthnConfigID identifies the authn provider to use to lookup users on the GitLab instance.
	// This should be the authn provider that's used to sign into the GitLab instance.
	AuthnConfigID providers.ConfigID

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

func newSudoProvider(op SudoProviderOp, cli httpcli.Doer) *SudoProvider {
	p := &SudoProvider{
		sudoToken: op.SudoToken,

		urn:               op.URN,
		clientProvider:    gitlab.NewClientProvider(op.BaseURL, cli),
		clientURL:         op.BaseURL,
		codeHost:          extsvc.NewCodeHost(op.BaseURL, extsvc.TypeGitLab),
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

func (p *SudoProvider) Validate() (problems []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, _, err := p.clientProvider.GetPATClient(p.sudoToken, "1").ListProjects(ctx, "projects"); err != nil {
		if err == ctx.Err() {
			problems = append(problems, fmt.Sprintf("GitLab API did not respond within 5s (%s)", err.Error()))
		} else if !gitlab.IsNotFound(err) {
			problems = append(problems, "access token did not have sufficient privileges, requires scopes \"sudo\" and \"api\"")
		}
	}
	return problems
}

func (p *SudoProvider) URN() string {
	return p.urn
}

func (p *SudoProvider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *SudoProvider) ServiceType() string {
	return p.codeHost.ServiceType
}

// fetchProjVis fetches a repository's visibility with usr's credentials. It returns:
// - whether the project is accessible to the user,
// - the visibility if the repo is accessible (otherwise this is empty),
// - whether the repository contents are accessible to usr, and
// - any error encountered in fetching (not including an error due to the repository not being visible);
//   if the error is non-nil, all other return values should be disregraded
func (p *SudoProvider) fetchProjVis(ctx context.Context, sudo string, projID int) (
	isAccessible bool, vis gitlab.Visibility, isContentAccessible bool, err error,
) {
	proj, err := p.clientProvider.GetPATClient(p.sudoToken, sudo).GetProject(ctx, gitlab.GetProjectOp{
		ID:       projID,
		CommonOp: gitlab.CommonOp{NoCache: true},
	})
	if err != nil {
		return false, "", false, err
	}

	if proj.Visibility == gitlab.Public {
		return true, proj.Visibility, true, nil
	}

	if sudo == "" {
		return false, proj.Visibility, false, nil
	}

	// At this point, sudo is non-nil *and* project visibility is internal or private

	if proj.Visibility == gitlab.Internal {
		// All authenticated users can read the contents of all internal/public projects
		// (https://docs.gitlab.com/ee/user/permissions.html).
		return true, proj.Visibility, true, nil
	}

	// If project visibility is private and it's accessible to user, we still need to check if the user
	// can read the repository contents (i.e., does not merely have "Guest" permissions).

	if _, err := p.clientProvider.GetPATClient(p.sudoToken, sudo).ListTree(ctx, gitlab.ListTreeOp{
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

// FetchAccount satisfies the authz.Provider interface. It iterates through the current list of
// linked external accounts, find the one (if it exists) that matches the authn provider specified
// in the SudoProvider struct, and fetches the user account from the GitLab API using that identity.
func (p *SudoProvider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account) (mine *extsvc.Account, err error) {
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
		var authnAcct *extsvc.Account
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

	var accountData extsvc.AccountData
	gitlab.SetExternalAccountData(&accountData, glUser, nil)

	glExternalAccount := extsvc.Account{
		UserID: user.ID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.codeHost.ServiceType,
			ServiceID:   p.codeHost.ServiceID,
			AccountID:   strconv.Itoa(int(glUser.ID)),
		},
		AccountData: accountData,
	}
	return &glExternalAccount, nil
}

func (p *SudoProvider) fetchAccountByExternalUID(ctx context.Context, uid string) (*gitlab.User, error) {
	q := make(url.Values)
	q.Add("extern_uid", uid)
	q.Add("provider", p.gitlabProvider)
	q.Add("per_page", "2")
	glUsers, _, err := p.clientProvider.GetPATClient(p.sudoToken, "").ListUsers(ctx, "users?"+q.Encode())
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

func (p *SudoProvider) fetchAccountByUsername(ctx context.Context, username string) (*gitlab.User, error) {
	q := make(url.Values)
	q.Add("username", username)
	q.Add("per_page", "2")
	glUsers, _, err := p.clientProvider.GetPATClient(p.sudoToken, "").ListUsers(ctx, "users?"+q.Encode())
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

var getProviderByConfigID = providers.GetProviderByConfigID
