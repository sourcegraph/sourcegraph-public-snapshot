package authz

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	permgl "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	db.ExternalServices.GitLabValidators = append(db.ExternalServices.GitLabValidators, validateGitLabProvider)
}

func validateGitLabProvider(g *schema.GitLabConnection) error {
	_, err := gitlabProvider(conf.Get(), g)
	return err
}

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

	op := permgl.GitLabAuthzProviderOp{
		BaseURL:   glURL,
		SudoToken: gl.Token,
		AuthnConfigID: auth.ProviderConfigID{
			ID:   gl.Authorization.AuthnProvider.ConfigID,
			Type: gl.Authorization.AuthnProvider.Type,
		},
		GitLabProvider: gl.Authorization.AuthnProvider.GitlabProvider,
		CacheTTL:       ttl,
		MockCache:      nil,
	}
	if gl.Authorization.AuthnProvider.ConfigID == "" {
		// Note: In the future when we support sign-in via GitLab, we can check if that is
		// enabled and instead fall back to that.
		if env.InsecureDev {
			log15.Warn("Using username matching for debugging purposes, because `authz.authnProvider.configID` in the config was empty. This should ONLY occur in a development build.")
			op.UseNativeUsername = true
		} else {
			return nil, errors.New("`authz.authnProvider.configID` was empty. No users will be granted access to these repositories.")
		}
	} else if gl.Authorization.AuthnProvider.Type == "" {
		return nil, errors.New("`authz.authnProvider.type` was not specified, which means GitLab users cannot be resolved.")
	} else if gl.Authorization.AuthnProvider.GitlabProvider == "" {
		return nil, errors.New("`authz.authnProvider.gitlabProvider` was not specified, which means GitLab users cannot be resolved.")
	} else {
		// Best-effort determine if the authz.authnConfigID field refers to an item in auth.provider
		found := false
		for _, p := range cfg.Critical.AuthProviders {
			if p.Openidconnect != nil && p.Openidconnect.ConfigID == gl.Authorization.AuthnProvider.ConfigID && p.Openidconnect.Type == gl.Authorization.AuthnProvider.Type {
				found = true
				break
			}
			if p.Saml != nil && p.Saml.ConfigID == gl.Authorization.AuthnProvider.ConfigID && p.Saml.Type == gl.Authorization.AuthnProvider.Type {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("Could not find item in `auth.providers` with config ID %q and type %q", gl.Authorization.AuthnProvider.ConfigID, gl.Authorization.AuthnProvider.Type)
		}
	}

	return NewGitLabProvider(op), nil
}

// NewGitLabProvider is a mockable constructor for new GitLabAuthzProvider instances.
var NewGitLabProvider = func(op permgl.GitLabAuthzProviderOp) authz.Provider {
	return permgl.NewProvider(op)
}
