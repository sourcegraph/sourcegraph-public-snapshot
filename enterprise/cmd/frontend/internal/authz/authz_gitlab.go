package authz

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	permgl "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func gitlabProviders(ctx context.Context, cfg *conf.Unified) (
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	gitlabs, err := db.ExternalServices.ListGitLabConnections(ctx)
	if err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load GitLab external service configs: %s", err))
		return
	}

	// Authorization (i.e., permissions) providers
	for _, gl := range gitlabs {
		p, err := gitlabProvider(cfg, gl)
		if err != nil {
			seriousProblems = append(seriousProblems, err.Error())
			continue
		}
		if p != nil {
			authzProviders = append(authzProviders, p)
		}
	}
	for _, provider := range authzProviders {
		for _, problem := range provider.Validate() {
			warnings = append(warnings, fmt.Sprintf("GitLab config for %s was invalid: %s", provider.ServiceID(), problem))
		}
	}
	return authzProviders, seriousProblems, warnings
}

func gitlabProvider(cfg *conf.Unified, gl *schema.GitLabConnection) (authz.Provider, error) {
	if gl.Authorization == nil {
		return nil, nil
	}

	glURL, err := url.Parse(gl.Url)
	if err != nil {
		return nil, fmt.Errorf("Could not parse URL for GitLab instance %q: %s", gl.Url, err)
	}

	ttl, err := parseTTL(gl.Authorization.Ttl)
	if err != nil {
		return nil, err
	}

	switch idp := gl.Authorization.IdentityProvider; {
	case idp.Oauth != nil:
		// Check that there is a GitLab authn provider corresponding to this GitLab instance
		foundAuthProvider := false
		for _, authnProvider := range cfg.Critical.AuthProviders {
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
			return nil, fmt.Errorf("Did not find authentication provider matching %q", gl.Url)
		}

		return NewGitLabOAuthProvider(permgl.GitLabOAuthAuthzProviderOp{
			BaseURL:  glURL,
			CacheTTL: ttl,
		}), nil
	case idp.Username != nil:
		return NewGitLabSudoProvider(permgl.SudoProviderOp{
			BaseURL:           glURL,
			SudoToken:         gl.Token,
			CacheTTL:          ttl,
			UseNativeUsername: true,
		}), nil
	case idp.External != nil:
		ext := idp.External
		for _, authProvider := range cfg.Critical.AuthProviders {
			saml := authProvider.Saml
			foundMatchingSAML := (saml != nil && saml.ConfigID == ext.AuthProviderID && ext.AuthProviderType == saml.Type)
			oidc := authProvider.Openidconnect
			foundMatchingOIDC := (oidc != nil && oidc.ConfigID == ext.AuthProviderID && ext.AuthProviderType == oidc.Type)
			if foundMatchingSAML || foundMatchingOIDC {
				return NewGitLabSudoProvider(permgl.SudoProviderOp{
					BaseURL: glURL,
					AuthnConfigID: auth.ProviderConfigID{
						Type: ext.AuthProviderType,
						ID:   ext.AuthProviderID,
					},
					GitLabProvider:    ext.GitlabProvider,
					SudoToken:         gl.Token,
					CacheTTL:          ttl,
					UseNativeUsername: false,
				}), nil
			}
		}
		return nil, fmt.Errorf("Did not find authentication provider matching type %s and configID %s", ext.AuthProviderType, ext.AuthProviderID)
	default:
		return nil, fmt.Errorf("No identityProvider was specified")
	}
}

// NewGitLabOAuthProvider is a mockable constructor for new gitlab.GitLabOAuthAuthzProvider instances.
var NewGitLabOAuthProvider = func(op permgl.GitLabOAuthAuthzProviderOp) authz.Provider {
	return permgl.NewOAuthProvider(op)
}

// NewGitLabSudoProvider is a mockable constructor for new gitlab.SudoProvider instances
var NewGitLabSudoProvider = func(op permgl.SudoProviderOp) authz.Provider {
	return permgl.NewSudoProvider(op)
}
