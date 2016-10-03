package backend

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestDefsService_ListRefs(t *testing.T) {
	var s defs
	ctx := testContext()

	want := []*graph.Ref{{File: "f"}}

	calledReposGet := Mocks.Repos.MockGet_Path(t, 1, "r")
	calledRefs := localstore.GraphMockRefs(&localstore.Mocks.Graph, want...)

	refs, err := s.ListRefs(ctx, &sourcegraph.DefsListRefsOp{Def: sourcegraph.DefSpec{CommitID: "c", Repo: 1, Path: "p"}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(refs.Refs, want) {
		t.Errorf("got %+v, want %+v", refs.Refs, want)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledRefs {
		t.Error("!calledRefs")
	}
}
