package bitbucketcloud

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/licensing"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	atypes "github.com/sourcegraph/sourcegraph/internal/authz/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of Bitbucket Cloud authz providers derived from the connections.
//
// It also returns any simple validation problems with the config, separating these into "serious problems"
// and "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
//
// This constructor does not and should not directly check connectivity to external services - if
// desired, callers should use `(*Provider).ValidateConnection` directly to get warnings related
// to connection issues.
func NewAuthzProviders(db database.DB, conns []*types.BitbucketCloudConnection, authProviders []schema.AuthProviders) *atypes.ProviderInitResult {
	initResults := &atypes.ProviderInitResult{}
	bbcloudAuthProviders := make(map[string]*schema.BitbucketCloudAuthProvider)
	for _, p := range authProviders {
		if p.Bitbucketcloud != nil {
			var id string
			bbURL, err := url.Parse(p.Bitbucketcloud.GetURL())
			if err != nil {
				// error reporting for this should happen elsewhere, for now just use what is given
				id = p.Bitbucketcloud.GetURL()
			} else {
				// use codehost normalized URL as ID
				ch := extsvc.NewCodeHost(bbURL, p.Bitbucketcloud.Type)
				id = ch.ServiceID
			}
			bbcloudAuthProviders[id] = p.Bitbucketcloud
		}
	}

	for _, c := range conns {
		p, err := newAuthzProvider(db, c)
		if err != nil {
			initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeBitbucketCloud)
			initResults.Problems = append(initResults.Problems, err.Error())
		}
		if p == nil {
			continue
		}

		initResults.Providers = append(initResults.Providers, p)
	}

	return initResults
}

func newAuthzProvider(
	db database.DB,
	c *types.BitbucketCloudConnection,
) (authz.Provider, error) {
	// If authorization is not set for this connection, we do not need an
	// authz provider.
	if c.Authorization == nil {
		return nil, nil
	}
	if err := licensing.Check(licensing.FeatureACLs); err != nil {
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

// ValidateAuthz validates the authorization fields of the given Perforce
// external service config.
func ValidateAuthz(_ *schema.BitbucketCloudConnection) error {
	// newAuthzProvider always succeeds, so directly return nil here.
	return nil
}
