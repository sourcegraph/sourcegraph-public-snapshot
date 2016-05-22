package mock

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func (s *DefsServer) MockGet(t *testing.T, wantDef sourcegraph.DefSpec) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		*called = true
		def := op.Def
		if def != wantDef {
			t.Errorf("got def %+v, want %+v", def, wantDef)
			return nil, grpc.Errorf(codes.NotFound, "def %v not found", wantDef)
		}
		return &sourcegraph.Def{Def: graph.Def{DefKey: def.DefKey()}}, nil
	}
	return
}

func (s *DefsServer) MockGet_Return(t *testing.T, wantDef *sourcegraph.Def) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		*called = true
		def := op.Def
		if def != wantDef.DefSpec() {
			t.Errorf("got def %+v, want %+v", def, wantDef.DefSpec())
			return nil, grpc.Errorf(codes.NotFound, "def %v not found", wantDef.DefKey)
		}
		return wantDef, nil
	}
	return
}

func (s *DefsServer) MockList(t *testing.T, wantDefs ...*sourcegraph.Def) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, opt *sourcegraph.DefListOptions) (*sourcegraph.DefList, error) {
		*called = true
		return &sourcegraph.DefList{Defs: wantDefs}, nil
	}
	return
}
