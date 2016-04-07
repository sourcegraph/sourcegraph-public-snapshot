package localstore

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
	"sourcegraph.com/sourcegraph/srclib/unit"

	"golang.org/x/net/context"
)

func (g *globalRefs) mustUpdate(ctx context.Context, t *testing.T, repo, unitName, unitType string, refs []*graph.Ref) error {
	op := &pb.ImportOp{
		Repo: repo,
		Unit: &unit.RepoSourceUnit{Unit: unitName, UnitType: unitType},
		Data: &graph.Output{
			Refs: refs,
		},
	}
	return g.Update(ctx, op)
}
