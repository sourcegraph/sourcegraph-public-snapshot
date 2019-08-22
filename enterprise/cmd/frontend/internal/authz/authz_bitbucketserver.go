package authz

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	bbsauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/schema"
)

func bitbucketServerProviders(
	ctx context.Context,
	db *sql.DB,
	cfg *conf.Unified,
	conns []*schema.BitbucketServerConnection,
) (
	authzProviders []authz.Provider,
	seriousProblems []string,
	warnings []string,
) {
	// Authorization (i.e., permissions) providers
	for _, c := range conns {
		if p, err := bitbucketServerProvider(db, c.Authorization, c.Url, c.Username, cfg.Critical.AuthProviders); err != nil {
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
	db *sql.DB,
	a *schema.BitbucketServerAuthorization,
	instanceURL, username string,
	ps []schema.AuthProviders,
) (authz.Provider, error) {
	if a == nil {
		return nil, nil
	}

	errs := new(multierror.Error)

	ttl, err := parseTTL(a.Ttl)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	hardTTL, err := parseTTL(a.HardTTL)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	if hardTTL < ttl {
		errs = multierror.Append(errs, errors.Errorf("authorization.hardTTL: must be larger than ttl"))
	}

	baseURL, err := url.Parse(instanceURL)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	cli := bitbucketserver.NewClient(baseURL, nil)
	cli.Username = username

	if err = cli.SetOAuth(a.Oauth.ConsumerKey, a.Oauth.SigningKey); err != nil {
		errs = multierror.Append(errs, errors.Wrap(err, "authorization.oauth.signingKey"))
	}

	var p authz.Provider
	switch idp := a.IdentityProvider; {
	case idp.Username != nil:
		p = bbsauthz.NewProvider(cli, db, ttl, hardTTL)
	default:
		errs = multierror.Append(errs, errors.Errorf("No identityProvider was specified"))
	}

	return p, errs.ErrorOrNil()
}

// ValidateBitbucketServerAuthz validates the authorization fields of the given BitbucketServer external
// service config.
func ValidateBitbucketServerAuthz(c *schema.BitbucketServerConnection, ps []schema.AuthProviders) error {
	_, err := bitbucketServerProvider(nil, c.Authorization, c.Url, c.Username, ps)
	return err
}
