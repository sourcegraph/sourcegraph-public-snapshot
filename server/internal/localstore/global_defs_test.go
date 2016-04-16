package localstore

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
	"sourcegraph.com/sourcegraph/srclib/unit"

	"golang.org/x/net/context"
)

func (g *globalDefs) mustUpdate(ctx context.Context, t *testing.T, repo, unitName, unitType string, defs []*graph.Def) error {
	op := &pb.ImportOp{
		Repo: repo,
		Unit: &unit.RepoSourceUnit{Unit: unitName, UnitType: unitType},
		Data: &graph.Output{
			Defs: defs,
		},
	}
	return g.Update(ctx, op)
}
