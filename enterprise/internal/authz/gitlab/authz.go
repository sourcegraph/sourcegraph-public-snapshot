package gitlab

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
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

type ExternalConnection struct {
	*types.ExternalService
	*types.GitLabConnection
}

func NewAuthzProviders(
	cfg schema.SiteConfiguration,
	conns []*ExternalConnection,
) (ps []authz.Provider, problems []string, warnings []string) {
	// Authorization (i.e., permissions) providers

	for _, c := range conns {
		p, err := newAuthzProvider(c, gitlab.TokenType(c.TokenType), cfg.AuthProviders)
		if err != nil {
			problems = append(problems, err.Error())
		} else if p != nil {
			ps = append(ps, p)
		}
	}

	return ps, problems, warnings
}

func newAuthzProvider(c *ExternalConnection, tokenType gitlab.TokenType, ps []schema.AuthProviders) (authz.Provider, error) {
	if c.GitLabConnection.Authorization == nil {
		return nil, nil
	}

	glURL, err := url.Parse(c.GitLabConnection.Url)
	if err != nil {
		return nil, errors.Errorf("Could not parse URL for GitLab instance %q: %s", c.GitLabConnection.Url, err)
	}

	switch idp := c.GitLabConnection.Authorization.IdentityProvider; {
	case idp.Oauth != nil:
		// Check that there is a GitLab authn provider corresponding to this GitLab instance
		foundAuthProvider := false

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
				break
			}
		}
		if !foundAuthProvider {
			return nil, errors.Errorf("Did not find authentication provider matching %q. Check the [**site configuration**](/site-admin/configuration) to verify an entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) exists for %s.", c.GitLabConnection.Url, c.GitLabConnection.Url)
		}

		return NewOAuthProvider(OAuthProviderOp{
			URN:             c.GitLabConnection.URN,
			BaseURL:         glURL,
			Token:           c.GitLabConnection.Token,
			TokenType:       tokenType,
			ExternalService: c.ExternalService,
		}), nil
	case idp.Username != nil:
		return NewSudoProvider(SudoProviderOp{
			URN:               c.GitLabConnection.URN,
			BaseURL:           glURL,
			SudoToken:         c.GitLabConnection.Token,
			UseNativeUsername: true,
		}), nil
	case idp.External != nil:
		ext := idp.External
		for _, authProvider := range ps {
			saml := authProvider.Saml
			foundMatchingSAML := saml != nil && saml.ConfigID == ext.AuthProviderID && ext.AuthProviderType == saml.Type
			oidc := authProvider.Openidconnect
			foundMatchingOIDC := oidc != nil && oidc.ConfigID == ext.AuthProviderID && ext.AuthProviderType == oidc.Type
			if foundMatchingSAML || foundMatchingOIDC {
				return NewSudoProvider(SudoProviderOp{
					URN:     c.GitLabConnection.URN,
					BaseURL: glURL,
					AuthnConfigID: providers.ConfigID{
						Type: ext.AuthProviderType,
						ID:   ext.AuthProviderID,
					},
					GitLabProvider:    ext.GitlabProvider,
					SudoToken:         c.GitLabConnection.Token,
					UseNativeUsername: false,
				}), nil
			}
		}
		return nil, errors.Errorf("Did not find authentication provider matching type %s and configID %s. Check the [**site configuration**](/site-admin/configuration) to verify that an entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) matches the type and configID.", ext.AuthProviderType, ext.AuthProviderID)
	default:
		return nil, errors.Errorf("No identityProvider was specified")
	}
}

// NewOAuthProvider is a mockable constructor for new OAuthProvider instances.
var NewOAuthProvider = func(op OAuthProviderOp) authz.Provider {
	var refreshToken string
	if op.TokenType == gitlab.TokenTypeOAuth {
		refreshToken = op.Token
	}

	helper := &database.RefreshTokenHelperForExternalService{
		DB:                op.db,
		ExternalServiceID: op.ExternalService.ID,
		OauthRefreshToken: refreshToken,
	}

	return newOAuthProvider(op, nil, helper.RefreshToken)
}

// NewSudoProvider is a mockable constructor for new SudoProvider instances.
var NewSudoProvider = func(op SudoProviderOp) authz.Provider {
	return newSudoProvider(op, nil)
}

// ValidateAuthz validates the authorization fields of the given GitLab external
// service config.
func ValidateAuthz(cfg *types.GitLabConnection, ps []schema.AuthProviders) error {
	_, err := newAuthzProvider(&ExternalConnection{GitLabConnection: cfg}, gitlab.TokenType(cfg.TokenType), ps)
	return err
}
