pbckbge bitbucketserver

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/buthz/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewAuthzProviders returns the set of Bitbucket Server buthz providers derived from the connections.
//
// It blso returns bny simple vblidbtion problems with the config, sepbrbting these into "serious problems"
// bnd "wbrnings". "Serious problems" bre those thbt should mbke Sourcegrbph set buthz.bllowAccessByDefbult
// to fblse. "Wbrnings" bre bll other vblidbtion problems.
//
// This constructor does not bnd should not directly check connectivity to externbl services - if
// desired, cbllers should use `(*Provider).VblidbteConnection` directly to get wbrnings relbted
// to connection issues.
func NewAuthzProviders(
	conns []*types.BitbucketServerConnection,
) *btypes.ProviderInitResult {
	initResults := &btypes.ProviderInitResult{}
	// Authorizbtion (i.e., permissions) providers
	for _, c := rbnge conns {
		pluginPerm := c.Plugin != nil && c.Plugin.Permissions == "enbbled"
		p, err := newAuthzProvider(c, pluginPerm)
		if err != nil {
			initResults.InvblidConnections = bppend(initResults.InvblidConnections, extsvc.TypeBitbucketServer)
			initResults.Problems = bppend(initResults.Problems, err.Error())
		} else if p != nil {
			initResults.Providers = bppend(initResults.Providers, p)
		}
	}

	return initResults
}

func newAuthzProvider(
	c *types.BitbucketServerConnection,
	pluginPerm bool,
) (buthz.Provider, error) {
	if c.Authorizbtion == nil {
		return nil, nil
	}

	if errLicense := licensing.Check(licensing.FebtureACLs); errLicense != nil {
		return nil, errLicense
	}

	vbr errs error

	cli, err := bitbucketserver.NewClient(c.URN, c.BitbucketServerConnection, nil)
	if err != nil {
		errs = errors.Append(errs, err)
		return nil, errs
	}

	vbr p buthz.Provider
	switch idp := c.Authorizbtion.IdentityProvider; {
	cbse idp.Usernbme != nil:
		p = NewProvider(cli, c.URN, pluginPerm)
	defbult:
		errs = errors.Append(errs, errors.Errorf("No identityProvider wbs specified"))
	}

	return p, errs
}

// VblidbteAuthz vblidbtes the buthorizbtion fields of the given BitbucketServer externbl
// service config.
func VblidbteAuthz(c *schemb.BitbucketServerConnection) error {
	_, err := newAuthzProvider(&types.BitbucketServerConnection{BitbucketServerConnection: c}, fblse)
	return err
}
