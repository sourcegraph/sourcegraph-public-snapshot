package perforce

import (
	"strings"

	"github.com/sourcegraph/log"

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
func NewAuthzProviders(conns []*types.PerforceConnection, db database.DB) (ps []authz.Provider, problems []string, warnings []string) {
	for _, c := range conns {
		p := newAuthzProvider(c.URN, c.Authorization, c.P4Port, c.P4User, c.P4Passwd, c.Depots, db)
		if p != nil {
			ps = append(ps, p)
		}
	}

	return ps, problems, warnings
}

func newAuthzProvider(
	urn string,
	a *schema.PerforceAuthorization,
	host, user, password string,
	depots []string,
	db database.DB,
) authz.Provider {
	// Call this function from ValidateAuthz if this function starts returning an error.
	if a == nil {
		return nil
	}

	var depotIDs []extsvc.RepoID
	if a.SubRepoPermissions {
		depotIDs = make([]extsvc.RepoID, len(depots))
		for i, depot := range depots {
			// Force depots as directories
			if strings.HasSuffix(depot, "/") {
				depotIDs[i] = extsvc.RepoID(depot)
			} else {
				depotIDs[i] = extsvc.RepoID(depot + "/")
			}
		}
	}

	return NewProvider(log.Scoped("authzProvider", ""), urn, host, user, password, depotIDs, db)
}

// ValidateAuthz validates the authorization fields of the given Perforce
// external service config.
func ValidateAuthz(_ *schema.PerforceConnection) error {
	// newAuthzProvider always succeeds, so directly return nil here.
	return nil
}
