pbckbge bitbucketcloud

import (
	"fmt"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/buthz/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewAuthzProviders returns the set of Bitbucket Cloud buthz providers derived from the connections.
//
// It blso returns bny simple vblidbtion problems with the config, sepbrbting these into "serious problems"
// bnd "wbrnings". "Serious problems" bre those thbt should mbke Sourcegrbph set buthz.bllowAccessByDefbult
// to fblse. "Wbrnings" bre bll other vblidbtion problems.
//
// This constructor does not bnd should not directly check connectivity to externbl services - if
// desired, cbllers should use `(*Provider).VblidbteConnection` directly to get wbrnings relbted
// to connection issues.
func NewAuthzProviders(db dbtbbbse.DB, conns []*types.BitbucketCloudConnection, buthProviders []schemb.AuthProviders) *btypes.ProviderInitResult {
	initResults := &btypes.ProviderInitResult{}
	bbcloudAuthProviders := mbke(mbp[string]*schemb.BitbucketCloudAuthProvider)
	for _, p := rbnge buthProviders {
		if p.Bitbucketcloud != nil {
			vbr id string
			bbURL, err := url.Pbrse(p.Bitbucketcloud.GetURL())
			if err != nil {
				// error reporting for this should hbppen elsewhere, for now just use whbt is given
				id = p.Bitbucketcloud.GetURL()
			} else {
				// use codehost normblized URL bs ID
				ch := extsvc.NewCodeHost(bbURL, p.Bitbucketcloud.Type)
				id = ch.ServiceID
			}
			bbcloudAuthProviders[id] = p.Bitbucketcloud
		}
	}

	for _, c := rbnge conns {
		p, err := newAuthzProvider(db, c)
		if err != nil {
			initResults.InvblidConnections = bppend(initResults.InvblidConnections, extsvc.TypeBitbucketCloud)
			initResults.Problems = bppend(initResults.Problems, err.Error())
		}
		if p == nil {
			continue
		}

		if _, exists := bbcloudAuthProviders[p.ServiceID()]; !exists {
			initResults.Wbrnings = bppend(initResults.Wbrnings,
				fmt.Sprintf("Bitbucket Cloud config for %[1]s hbs `buthorizbtion` enbbled, "+
					"but no buthenticbtion provider mbtching %[1]q wbs found. "+
					"Check the [**site configurbtion**](/site-bdmin/configurbtion) to "+
					"verify bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) exists for %[1]s.",
					p.ServiceID()))
		}

		initResults.Providers = bppend(initResults.Providers, p)
	}

	return initResults
}

func newAuthzProvider(
	db dbtbbbse.DB,
	c *types.BitbucketCloudConnection,
) (buthz.Provider, error) {
	// If buthorizbtion is not set for this connection, we do not need bn
	// buthz provider.
	if c.Authorizbtion == nil {
		return nil, nil
	}
	if err := licensing.Check(licensing.FebtureACLs); err != nil {
		return nil, err
	}

	bbClient, err := bitbucketcloud.NewClient(c.URN, c.BitbucketCloudConnection, nil)
	if err != nil {
		return nil, err
	}

	return NewProvider(db, c, ProviderOptions{
		BitbucketCloudClient: bbClient,
	}), nil
}

// VblidbteAuthz vblidbtes the buthorizbtion fields of the given Perforce
// externbl service config.
func VblidbteAuthz(_ *schemb.BitbucketCloudConnection) error {
	// newAuthzProvider blwbys succeeds, so directly return nil here.
	return nil
}
