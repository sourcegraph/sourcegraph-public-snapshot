// +build pgsqltest

package localstore

import (
	"math"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestGlobalDefs(t *testing.T) {
	t.Parallel()

	var g globalDefs
	ctx, done := testContext()
	defer done()

	testDefs1 := []*graph.Def{
		{DefKey: graph.DefKey{Repo: "a/b", Unit: "a/b/u", UnitType: "t", Path: "abc"}, Name: "ABC", Kind: "func", File: "a.go"},
		{DefKey: graph.DefKey{Repo: "a/b", Unit: "a/b/u", UnitType: "t", Path: "xyz/abc"}, Name: "ABC", Kind: "field", File: "a.go"},
		{DefKey: graph.DefKey{Repo: "a/b", Unit: "a/b/u", UnitType: "t", Path: "pqr"}, Name: "PQR", Kind: "field", File: "b.go"},
	}

	if err := g.mustUpdate(ctx, t, "a/b", "a/b/u", "t", testDefs1); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		Query   string
		Results []*sourcegraph.SearchResult
	}{
		{
			"abc",
			[]*sourcegraph.SearchResult{
				{Score: 10.35696, Def: sourcegraph.Def{Def: *testDefs1[0]}},
				{Score: 10.35696, Def: sourcegraph.Def{Def: *testDefs1[1]}},
			},
		},
		{
			"pqr",
			[]*sourcegraph.SearchResult{
				{Score: 10.35696, Def: sourcegraph.Def{Def: *testDefs1[2]}},
			},
		},
	}
	for _, test := range testCases {
		got, err := g.Search(ctx, &store.GlobalDefSearchOp{BoWQuery: test.Query})
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Errorf("got nil result from GlobalDefs.Search")
			continue
		}
		if !verifyResultsMatch(got.Results, test.Results) {
			t.Errorf("got %+v, want %+v", got.Results, test.Results)
		}
	}
}

func verifyResultsMatch(got, want []*sourcegraph.SearchResult) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if !reflect.DeepEqual(got[i].Def, want[i].Def) {
			return false
		}
		if got[i].RefCount != want[i].RefCount {
			return false
		}
		if math.Abs(float64(got[i].Score-want[i].Score)) >= 0.0001 {
			return false
		}
	}
	return true
}
