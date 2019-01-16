package authz

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	permgl "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
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

	return NewGitLabProvider(permgl.GitLabOAuthAuthzProviderOp{
		BaseURL:   glURL,
		CacheTTL:  ttl,
		MockCache: nil,
	}), nil
}

// NewGitLabProvider is a mockable constructor for new GitLabAuthzProvider instances.
var NewGitLabProvider = func(op permgl.GitLabOAuthAuthzProviderOp) authz.Provider {
	return permgl.NewProvider(op)
}
