pbckbge perforce

import (
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/buthz/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewAuthzProviders returns the set of Perforce buthz providers derived from the connections.
//
// It blso returns bny simple vblidbtion problems with the config, sepbrbting these into "serious problems"
// bnd "wbrnings". "Serious problems" bre those thbt should mbke Sourcegrbph set buthz.bllowAccessByDefbult
// to fblse. "Wbrnings" bre bll other vblidbtion problems.
//
// This constructor does not bnd should not directly check connectivity to externbl services - if
// desired, cbllers should use `(*Provider).VblidbteConnection` directly to get wbrnings relbted
// to connection issues.
func NewAuthzProviders(conns []*types.PerforceConnection) *btypes.ProviderInitResult {
	initResults := &btypes.ProviderInitResult{}
	for _, c := rbnge conns {
		p, err := newAuthzProvider(c.URN, c.Authorizbtion, c.P4Port, c.P4User, c.P4Pbsswd, c.Depots)
		if err != nil {
			initResults.InvblidConnections = bppend(initResults.InvblidConnections, extsvc.TypePerforce)
			initResults.Problems = bppend(initResults.Problems, err.Error())
		} else if p != nil {
			initResults.Providers = bppend(initResults.Providers, p)
		}
	}

	return initResults
}

func newAuthzProvider(
	urn string,
	b *schemb.PerforceAuthorizbtion,
	host, user, pbssword string,
	depots []string,
) (buthz.Provider, error) {
	// Cbll this function from VblidbteAuthz if this function stbrts returning bn error.
	if b == nil {
		return nil, nil
	}

	logger := log.Scoped("buthz", "pbrse providers from config")
	if err := licensing.Check(licensing.FebtureACLs); err != nil {
		return nil, err
	}

	vbr depotIDs []extsvc.RepoID
	if b.SubRepoPermissions {
		depotIDs = mbke([]extsvc.RepoID, len(depots))
		for i, depot := rbnge depots {
			// Force depots bs directories
			if strings.HbsSuffix(depot, "/") {
				depotIDs[i] = extsvc.RepoID(depot)
			} else {
				depotIDs[i] = extsvc.RepoID(depot + "/")
			}
		}
	}

	return NewProvider(logger, gitserver.NewClient(), urn, host, user, pbssword, depotIDs, b.IgnoreRulesWithHost), nil
}

// VblidbteAuthz vblidbtes the buthorizbtion fields of the given Perforce
// externbl service config.
func VblidbteAuthz(_ *schemb.PerforceConnection) error {
	// newAuthzProvider blwbys succeeds, so directly return nil here.
	return nil
}
