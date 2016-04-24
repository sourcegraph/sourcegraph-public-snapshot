package localstore

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"

	"golang.org/x/net/context"
)

func (g *globalDefs) mustUpdate(ctx context.Context, t *testing.T, repo, unitName, unitType string, defs []*graph.Def) error {
	// TODO: remove this entire method?

	// op := &pb.ImportOp{
	// 	Repo: repo,
	// 	Unit: &unit.RepoSourceUnit{Unit: unitName, UnitType: unitType},
	// 	Data: &graph.Output{
	// 		Defs: defs,
	// 	},
	// }

	// return g.Update(ctx, op)
	return nil
}
