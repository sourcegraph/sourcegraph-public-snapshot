package gitlab

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	atypes "github.com/sourcegraph/sourcegraph/internal/authz/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of GitLab authz providers derived from the connections.
//
// It also returns any simple validation problems with the config, separating these into "serious problems"
// and "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
//
// This constructor does not and should not directly check connectivity to external services - if
// desired, callers should use `(*Provider).ValidateConnection` directly to get warnings related
// to connection issues.
func NewAuthzProviders(
	db database.DB,
	conns []*types.GitLabConnection,
	ap []schema.AuthProviders,
) *atypes.ProviderInitResult {
	initResults := &atypes.ProviderInitResult{}
	// Authorization (i.e., permissions) providers
	for _, c := range conns {
		up, rp, err := newAuthzProvider(db, c, ap)
		if err != nil {
			initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeGitLab)
			initResults.Problems = append(initResults.Problems, err.Error())
			continue
		}
		if up != nil {
			initResults.UserPermissionsFetchers = append(initResults.UserPermissionsFetchers, up)
		}
		if rp != nil {
			initResults.RepoPermissionsFetchers = append(initResults.RepoPermissionsFetchers, rp)
		}
	}

	return initResults
}

func newAuthzProvider(db database.DB, c *types.GitLabConnection, ps []schema.AuthProviders) (authz.UserPermissionsFetcher, authz.RepoPermissionsFetcher, error) {
	if c.Authorization == nil {
		return nil, nil, nil
	}

	if errLicense := licensing.Check(licensing.FeatureACLs); errLicense != nil {
		return nil, nil, errLicense
	}

	glURL, err := url.Parse(c.Url)
	if err != nil {
		return nil, nil, errors.Errorf("Could not parse URL for GitLab instance %q: %s", c.Url, err)
	}

	switch idp := c.Authorization.IdentityProvider; {
	case idp.Oauth != nil:
		// Check that there is a GitLab authn provider corresponding to this GitLab instance
		foundAuthProvider := false
		syncInternalRepoPermissions := true
		for _, authnProvider := range ps {
			if authnProvider.Gitlab == nil {
				continue
			}
			authnURL := authnProvider.Gitlab.Url
			if authnURL == "" {
				authnURL = "https://gitlab.com"
			}
			authProviderURL, err := url.Parse(authnURL)
			if err != nil {
				// Ignore the error here, because the authn provider is responsible for its own validation
				continue
			}
			if authProviderURL.Hostname() == glURL.Hostname() {
				foundAuthProvider = true
				sirp := authnProvider.Gitlab.SyncInternalRepoPermissions
				syncInternalRepoPermissions = sirp == nil || *sirp
				break
			}
		}
		if !foundAuthProvider {
			return nil, nil, errors.Errorf("Did not find authentication provider matching %q. Check the [**site configuration**](/site-admin/configuration) to verify an entry in [`auth.providers`](https://sourcegraph.com/docs/admin/auth) exists for %s.", c.Url, c.Url)
		}

		p := newOAuthProvider(OAuthProviderOp{
			URN:                         c.URN,
			BaseURL:                     glURL,
			Token:                       c.Token,
			TokenType:                   gitlab.TokenType(c.TokenType),
			DB:                          db,
			SyncInternalRepoPermissions: syncInternalRepoPermissions,
		}, nil)
		return p, nil, nil
	case idp.Username != nil:
		p := newSudoProvider(SudoProviderOp{
			URN:                         c.URN,
			BaseURL:                     glURL,
			SudoToken:                   c.Token,
			SyncInternalRepoPermissions: !c.MarkInternalReposAsPublic,
		}, nil)
		return p, p, nil
	default:
		return nil, nil, errors.Errorf("No identityProvider was specified")
	}
}

// ValidateAuthz validates the authorization fields of the given GitLab external
// service config.
func ValidateAuthz(cfg *schema.GitLabConnection, ps []schema.AuthProviders) error {
	_, _, err := newAuthzProvider(nil, &types.GitLabConnection{GitLabConnection: cfg}, ps)
	return err
}
