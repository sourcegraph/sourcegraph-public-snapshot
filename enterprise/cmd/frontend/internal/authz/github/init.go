package github

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	iauthz.NewProviderRegister(func(
		ctx context.Context,
		cfg *conf.Unified,
		s iauthz.ExternalServicesStore,
		db *sql.DB,
	) (ps []authz.Provider, problems []string, warnings []string) {
		conns, err := s.ListGitHubConnections(ctx)
		if err != nil {
			problems = append(problems, fmt.Sprintf("Could not load GitHub external service configs: %s", err))
			return nil, problems, nil
		}

		// Authorization (i.e., permissions) providers
		for _, c := range conns {
			p, err := newAuthzProvider(c.Authorization, c.Url, c.Token)
			if err != nil {
				problems = append(problems, err.Error())
			} else if p != nil {
				ps = append(ps, p)
			}
		}

		for _, p := range ps {
			for _, problem := range p.Validate() {
				warnings = append(warnings, fmt.Sprintf("GitHub config for %s was invalid: %s", p.ServiceID(), problem))
			}
		}

		return ps, problems, warnings
	})
}

func newAuthzProvider(a *schema.GitHubAuthorization, instanceURL, token string) (authz.Provider, error) {
	if a == nil {
		return nil, nil
	}

	ghURL, err := url.Parse(instanceURL)
	if err != nil {
		return nil, fmt.Errorf("Could not parse URL for GitHub instance %q: %s", instanceURL, err)
	}

	ttl, err := iauthz.ParseTTL(a.Ttl)
	if err != nil {
		return nil, err
	}

	return NewProvider(ghURL, token, ttl, nil), nil
}

// ValidateGitHubAuthz validates the authorization fields of the given GitHub external
// service config.
func ValidateAuthz(cfg *schema.GitHubConnection) error {
	_, err := newAuthzProvider(cfg.Authorization, cfg.Url, cfg.Token)
	return err
}
