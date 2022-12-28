package bitbucketcloud

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"

	"github.com/ktrysmt/go-bitbucket"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of Perforce authz providers derived from the connections.
//
// It also returns any simple validation problems with the config, separating these into "serious problems"
// and "warnings". "Serious problems" are those that should make Sourcegraph set authz.allowAccessByDefault
// to false. "Warnings" are all other validation problems.
//
// This constructor does not and should not directly check connectivity to external services - if
// desired, callers should use `(*Provider).ValidateConnection` directly to get warnings related
// to connection issues.
func NewAuthzProviders(conns []*types.BitbucketCloudConnection, db database.DB) (ps []authz.Provider, problems []string, warnings []string, invalidConnections []string) {
	for _, c := range conns {
		bbURL, _ := url.Parse(c.Url)
		p, err := newAuthzProvider(c, bbURL, c.URN, c.Username, c.AppPassword)
		if err != nil {
			invalidConnections = append(invalidConnections, extsvc.TypePerforce)
			problems = append(problems, err.Error())
		} else if p != nil {
			ps = append(ps, p)
		}
	}

	return ps, problems, warnings, invalidConnections
}

func newAuthzProvider(
	c *types.BitbucketCloudConnection,
	url *url.URL,
	urn string,
	username string,
	appPassword string,
) (authz.Provider, error) {
	if c.Authorization == nil {
		return nil, nil
	}
	if err := licensing.Check(licensing.FeatureACLs); err != nil {
		return nil, err
	}

	bbClient := bitbucket.NewBasicAuth(username, appPassword)

	return NewProvider(url, urn, bbClient), nil
}

// ValidateAuthz validates the authorization fields of the given Perforce
// external service config.
func ValidateAuthz(_ *schema.BitbucketCloudConnection) error {
	// newAuthzProvider always succeeds, so directly return nil here.
	return nil
}
