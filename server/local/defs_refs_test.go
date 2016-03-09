package local

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/mockstore"
)

func TestDefsService_ListRefs(t *testing.T) {
	var s defs
	ctx, mock := testContext()

	want := []*graph.Ref{{File: "f"}}

	calledRefs := mockstore.GraphMockRefs(&mock.stores.Graph, want...)

	refs, err := s.ListRefs(ctx, &sourcegraph.DefsListRefsOp{Def: sourcegraph.DefSpec{CommitID: "c", Repo: "r", Path: "p"}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(refs.Refs, want) {
		t.Errorf("got %+v, want %+v", refs.Refs, want)
	}
	if !*calledRefs {
		t.Error("!calledRefs")
	}
}
