package bitbucketserver

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	atypes "github.com/sourcegraph/sourcegraph/internal/authz/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of Bitbucket Server authz providers derived from the connections.
//
// It also returns any simple validation problems with the config, separating these into "serious problems"
// and "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
//
// This constructor does not and should not directly check connectivity to external services - if
// desired, callers should use `(*Provider).ValidateConnection` directly to get warnings related
// to connection issues.
func NewAuthzProviders(
	conns []*types.BitbucketServerConnection,
	logger log.Logger,
) *atypes.ProviderInitResult {
	initResults := &atypes.ProviderInitResult{}
	// Authorization (i.e., permissions) providers
	for _, c := range conns {
		pluginPerm := c.Plugin != nil && c.Plugin.Permissions == "enabled"
		p, err := newAuthzProvider(c, pluginPerm, logger)
		if err != nil {
			initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeBitbucketServer)
			initResults.Problems = append(initResults.Problems, err.Error())
		} else if p != nil {
			initResults.Providers = append(initResults.Providers, p)
		}
	}

	return initResults
}

func newAuthzProvider(
	c *types.BitbucketServerConnection,
	pluginPerm bool,
	logger log.Logger,
) (authz.Provider, error) {
	if c.Authorization == nil {
		return nil, nil
	}

	if errLicense := licensing.Check(licensing.FeatureACLs); errLicense != nil {
		return nil, errLicense
	}

	var errs error

	logger = logger.Scoped("BitbucketServerAuthzProvider")
	cli, err := bitbucketserver.NewClient(c.URN, c.BitbucketServerConnection, nil, logger)
	if err != nil {
		errs = errors.Append(errs, err)
		return nil, errs
	}

	var p authz.Provider
	switch idp := c.Authorization.IdentityProvider; {
	case idp.Username != nil:
		p = NewProvider(cli, c.URN, pluginPerm)
	default:
		errs = errors.Append(errs, errors.Errorf("No identityProvider was specified"))
	}

	return p, errs
}

// ValidateAuthz validates the authorization fields of the given BitbucketServer external
// service config.
func ValidateAuthz(c *schema.BitbucketServerConnection, logger log.Logger) error {
	_, err := newAuthzProvider(&types.BitbucketServerConnection{BitbucketServerConnection: c}, false, logger)
	return err
}
