package localstore

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/experiment"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
)

// globalRefsExperiment runs two implementations of GlobalRefs
type globalRefsExperiment struct {
	A, B store.GlobalRefs
}

var globalRefsExp = &globalRefsExperiment{A: &globalRefsNew{}, B: &globalRefs{}}

func (g *globalRefsExperiment) Get(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error) {
	e := experiment.Perf{
		Name: "GlobalRefs.Get",
		B:    func() { g.B.Get(ctx, op) },
	}
	done := e.StartA()
	defer done()
	return g.A.Get(ctx, op)
}

func (g *globalRefsExperiment) Update(ctx context.Context, op *pb.ImportOp) error {
	e := experiment.Perf{
		Name: "GlobalRefs.Update",
		B:    func() { g.B.Update(ctx, op) },
	}
	done := e.StartA()
	defer done()
	return g.A.Update(ctx, op)
}
