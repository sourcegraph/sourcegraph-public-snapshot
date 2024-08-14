package gerrit

import (
	"github.com/sourcegraph/log"

	atypes "github.com/sourcegraph/sourcegraph/internal/authz/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// NewAuthzProviders returns the set of Gerrit authz providers derived from the connections.
func NewAuthzProviders(conns []*types.GerritConnection, logger log.Logger) *atypes.ProviderInitResult {
	initResults := &atypes.ProviderInitResult{}

	logger = logger.Scoped("GerritAuthzProvider")

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

		p, err := NewProvider(c, logger)
		if err != nil {
			initResults.InvalidConnections = append(initResults.InvalidConnections, extsvc.TypeGerrit)
			initResults.Problems = append(initResults.Problems, err.Error())
		}
		if p != nil {
			initResults.Providers = append(initResults.Providers, p)
		}
	}
	return initResults
}
