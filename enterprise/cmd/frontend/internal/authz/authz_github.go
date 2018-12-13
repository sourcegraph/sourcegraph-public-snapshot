package authz

import (
	"fmt"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	permgh "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func githubProvidersFromConfig(cfg *conf.Unified) (
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	for _, g := range cfg.Github {
		if g.Authorization == nil {
			continue
		}

		ghURL, err := url.Parse(g.Url)
		if err != nil {
			seriousProblems = append(seriousProblems, fmt.Sprintf("Could not parse URL for GitHub instance %q: %s", g.Url, err))
			continue // omit authz provider if could not parse URL
		}

		var ttl time.Duration
		ttl, warnings = parseTTLOrDefault(g.Authorization.Ttl, 3*time.Hour, warnings)

		authzProviders = append(authzProviders, permgh.NewProvider(ghURL, g.Token, ttl, nil))
	}
	return authzProviders, seriousProblems, warnings
}
