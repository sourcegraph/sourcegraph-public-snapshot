package gerrit

import (
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// NewAuthzProviders returns the set of Gerrit authz providers derived from the connections.
func NewAuthzProviders(conns []*types.GerritConnection) (ps []authz.Provider) {
	for _, c := range conns {
		p, _ := NewProvider(c)
		if p != nil {
			ps = append(ps, p)
		}
	}
	return ps
}
