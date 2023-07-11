package gerrit

import (
	"fmt"
	"net/url"

	atypes "github.com/sourcegraph/sourcegraph/enterprise/internal/authz/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of Gerrit authz providers derived from the connections.
func NewAuthzProviders(conns []*types.GerritConnection, authProviders []schema.AuthProviders) *atypes.ProviderInitResult {
	initResults := &atypes.ProviderInitResult{}
	gerritAuthProviders := make(map[string]*schema.GerritAuthProvider)
	for _, p := range authProviders {
		if p.Gerrit == nil {
			continue
		}

		gerritURL, err := url.Parse(p.Gerrit.Url)
		if err != nil {
			continue
		}

		// Use normalised base URL as ID.
		gerritAuthProviders[extsvc.NormalizeBaseURL(gerritURL).String()] = p.Gerrit
	}

	for _, c := range conns {
		if c.Authorization == nil {
			// No authorization required
			continue
		}
		if err := licensing.Check(licensing.FeatureACLs); err != nil {
			initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeGerrit)
			initResults.Problems = append(initResults.Problems, err.Error())
			continue
		}
		p, err := NewProvider(c)
		if err != nil {
			initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeGerrit)
			initResults.Problems = append(initResults.Problems, err.Error())
		}
		if p != nil {
			initResults.Providers = append(initResults.Providers, p)

			if _, exists := gerritAuthProviders[p.ServiceID()]; !exists {
				initResults.Warnings = append(initResults.Warnings,
					fmt.Sprintf("Gerrit config for %[1]s has `authorization` enabled, "+
						"but no authentication provider matching %[1]q was found. "+
						"Check the [**site configuration**](/site-admin/configuration) to "+
						"verify an entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) exists for %[1]s.",
						p.ServiceID()))
			}
		}
	}
	return initResults
}
