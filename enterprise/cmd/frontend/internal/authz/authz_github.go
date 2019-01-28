package authz

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	permgh "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

func githubProviders(ctx context.Context) (
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	githubs, err := db.ExternalServices.ListGitHubConnections(ctx)
	if err != nil {
		seriousProblems = append(seriousProblems, fmt.Sprintf("Could not load GitHub external service configs: %s", err))
		return
	}

	for _, g := range githubs {
		p, err := githubProvider(g)
		if err != nil {
			seriousProblems = append(seriousProblems, err.Error())
			continue
		}
		if p != nil {
			authzProviders = append(authzProviders, p)
		}
	}
	return authzProviders, seriousProblems, warnings
}

func githubProvider(g *schema.GitHubConnection) (authz.Provider, error) {
	if g.Authorization == nil {
		return nil, nil
	}

	ghURL, err := url.Parse(g.Url)
	if err != nil {
		return nil, fmt.Errorf("Could not parse URL for GitHub instance %q: %s", g.Url, err)
	}

	ttl, err := parseTTL(g.Authorization.Ttl)
	if err != nil {
		return nil, err
	}

	return permgh.NewProvider(ghURL, g.Token, ttl, nil), nil
}
