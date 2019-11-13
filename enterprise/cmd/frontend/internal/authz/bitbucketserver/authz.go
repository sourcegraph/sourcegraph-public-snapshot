package bitbucketserver

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of Bitbucket Server authz providers derived from the connections.
// It also returns any validation problems with the config, separating these into "serious problems" and
// "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
func NewAuthzProviders(
	conns []*schema.BitbucketServerConnection,
	db *sql.DB,
) (ps []authz.Provider, problems []string, warnings []string) {
	// Authorization (i.e., permissions) providers
	for _, c := range conns {
		p, err := newAuthzProvider(db, c.Authorization, c.Url, c.Username)
		if err != nil {
			problems = append(problems, err.Error())
		} else if p != nil {
			ps = append(ps, p)
		}
	}

	for _, p := range ps {
		for _, problem := range p.Validate() {
			warnings = append(warnings, fmt.Sprintf("BitbucketServer config for %s was invalid: %s", p.ServiceID(), problem))
		}
	}

	return ps, problems, warnings
}

func newAuthzProvider(
	db *sql.DB,
	a *schema.BitbucketServerAuthorization,
	instanceURL, username string,
) (authz.Provider, error) {
	if a == nil {
		return nil, nil
	}

	errs := new(multierror.Error)

	ttl, err := iauthz.ParseTTL(a.Ttl)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	hardTTL, err := iauthz.ParseTTL(a.HardTTL)
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
		p = NewProvider(cli, db, ttl, hardTTL)
	default:
		errs = multierror.Append(errs, errors.Errorf("No identityProvider was specified"))
	}

	return p, errs.ErrorOrNil()
}

// ValidateAuthz validates the authorization fields of the given BitbucketServer external
// service config.
func ValidateAuthz(c *schema.BitbucketServerConnection) error {
	_, err := newAuthzProvider(nil, c.Authorization, c.Url, c.Username)
	return err
}
