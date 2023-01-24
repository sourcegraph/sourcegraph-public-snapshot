package gerrit

import (
	"fmt"
	"net/url"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// NewAuthzProviders returns the set of Gerrit authz providers derived from the connections.
func NewAuthzProviders(conns []*types.GerritConnection, authProviders []schema.AuthProviders) (ps []authz.Provider, problems []string, warnings []string, invalidConnections []string) {
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
			invalidConnections = append(invalidConnections, extsvc.TypeGerrit)
			problems = append(problems, err.Error())
			continue
		}
		p, err := NewProvider(c)
		if err != nil {
			invalidConnections = append(invalidConnections, extsvc.TypeGerrit)
			problems = append(problems, err.Error())
		}
		if p != nil {
			ps = append(ps, p)

			if _, exists := gerritAuthProviders[p.ServiceID()]; !exists {
				warnings = append(warnings,
					fmt.Sprintf("Gerrit config for %[1]s has `authorization` enabled, "+
						"but no authentication provider matching %[1]q was found. "+
						"Check the [**site configuration**](/site-admin/configuration) to "+
						"verify an entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) exists for %[1]s.",
						p.ServiceID()))
			}
		}
	}
	return
}
