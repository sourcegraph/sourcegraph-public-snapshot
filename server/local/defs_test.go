package local

import (
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/mockstore"
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
