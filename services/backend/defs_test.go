package backend

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/store"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store/mockstore"
)

func TestDefsService_List_Repos(t *testing.T) {
	var s defs
	ctx, mock := testContext()

	calledDefs := mockstore.GraphMockDefs(&mock.stores.Graph)
	calledGetRepo := mock.servers.Repos.MockGet(t, "r")

	_, err := s.List(ctx, &sourcegraph.DefListOptions{
		RepoRevs: []string{"r@tttttttttttttttttttttttttttttttttttttttt"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledDefs {
		t.Error("!calledDefs")
	}
	if !*calledGetRepo {
		t.Error("!calledGetRepo")
	}
}

func TestDefListOptions_Offset(t *testing.T) {
	adjDefs := []*graph.Def{
		{
			DefKey: graph.DefKey{
				Path: "p/1",
			},
			Name:     "d1",
			DefStart: 10,
			DefEnd:   20,
			Data:     []byte(`{}`),
		},
		{
			DefKey: graph.DefKey{
				Path: "p/2",
			},
			Name:     "d1",
			DefStart: 10,
			DefEnd:   20,
			Data:     []byte(`{}`),
		},
		{
			DefKey: graph.DefKey{
				Path: "q",
			},
			Name:     "d2",
			DefStart: 10,
			DefEnd:   20,
			Data:     []byte(`{}`),
		},
		{
			DefKey: graph.DefKey{
				Path: "p",
			},
			Name:     "d1",
			DefStart: 40,
			DefEnd:   50,
			Data:     []byte(`{}`),
		},
	}

	opt := &sourcegraph.DefListOptions{
		Name:      adjDefs[0].Name,
		ByteStart: adjDefs[0].DefStart,
		ByteEnd:   adjDefs[0].DefEnd,
	}

	defs := store.DefFilters(defListOptionsFilters(opt)).SelectDefs(adjDefs...)

	wantDefs := []*graph.Def{adjDefs[0], adjDefs[1]}
	if !reflect.DeepEqual(defs, wantDefs) {
		t.Errorf("got %+v, want %+v", defs, wantDefs)
	}
}
