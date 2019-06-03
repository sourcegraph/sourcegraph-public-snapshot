package authz

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func bitbucketServerProviders(
	ctx context.Context,
	cfg *conf.Unified,
	conns []*schema.BitbucketServerConnection,
) (
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	// Authorization (i.e., permissions) providers
	for _, c := range conns {
		if p, err := bitbucketServerProvider(c.Authorization, c.Url, c.Token, cfg.Critical.AuthProviders); err != nil {
			seriousProblems = append(seriousProblems, err.Error())
		} else if p != nil {
			authzProviders = append(authzProviders, p)
		}
	}

	for _, p := range authzProviders {
		for _, problem := range p.Validate() {
			warnings = append(warnings, fmt.Sprintf("BitbucketServer config for %s was invalid: %s", p.ServiceID(), problem))
		}
	}

	return authzProviders, seriousProblems, warnings
}

func bitbucketServerProvider(
	a *schema.BitbucketServerAuthorization,
	instanceURL, token string,
	ps []schema.AuthProviders,
) (authz.Provider, error) {
	if a == nil {
		return nil, nil
	}

	_, err := url.Parse(instanceURL)
	if err != nil {
		return nil, fmt.Errorf("Could not parse URL for BitbucketServer instance %q: %s", instanceURL, err)
	}

	_, err = parseTTL(a.Ttl)
	if err != nil {
		return nil, err
	}

	switch idp := a.IdentityProvider; {
	case idp.Username != nil:
		return nil, nil // TODO
	default:
		return nil, fmt.Errorf("No identityProvider was specified")
	}
}

// ValidateBitbucketServerAuthz validates the authorization fields of the given BitbucketServer external
// service config.
func ValidateBitbucketServerAuthz(c *schema.BitbucketServerConnection, ps []schema.AuthProviders) error {
	_, err := bitbucketServerProvider(c.Authorization, c.Url, c.Token, ps)
	return err
}
