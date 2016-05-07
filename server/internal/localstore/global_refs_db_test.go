// +build pgsqltest

package localstore

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestGlobalRefs(t *testing.T) {
	t.Parallel()

	var g globalRefs
	ctx, _, done := testContext()
	defer done()

	testRefs1 := []*graph.Ref{
		{DefPath: ".", DefRepo: "", DefUnit: "", File: "a/b/u/s.go"},              // package ref
		{DefPath: "A/R", DefRepo: "", DefUnit: "", File: "a/b/u/s.go", Def: true}, // def ref
		{DefPath: "A/R", DefRepo: "", DefUnit: "", File: "a/b/u/s.go"},            // same unit
		{DefPath: "A/R", DefRepo: "", DefUnit: "", File: "a/b/u/s.go"},            // same unit, repeated
		{DefPath: "A/S", DefRepo: "", DefUnit: "a/b/p", File: "a/b/u/s.go"},       // same repo, different unit
		{DefPath: "X/Y", DefRepo: "x/y", DefUnit: "x/y/c", File: "a/b/u/s.go"},    // different repo
		{DefPath: "A/R", DefRepo: "x/y", DefUnit: "x/y/c", File: "a/b/u/s.go"},    // different repo
	}
	testRefs2 := []*graph.Ref{
		{DefPath: "P/Q", DefRepo: "", DefUnit: "", File: "a/b/p/t.go"},         // same unit
		{DefPath: "A/R", DefRepo: "", DefUnit: "a/b/u", File: "a/b/p/t.go"},    // same repo, different unit
		{DefPath: "B/S", DefRepo: "x/y", DefUnit: "x/y/c", File: "a/b/p/t.go"}, // different repo
	}
	testRefs3 := []*graph.Ref{
		{DefPath: "", DefRepo: "", DefUnit: "", File: "x/y/c/v.go"},         // package ref
		{DefPath: "A/R", DefRepo: "", DefUnit: "x/y/c", File: "x/y/c/v.go"}, // same unit
		{DefPath: "B/T", DefRepo: "", DefUnit: "x/y/d", File: "x/y/c/v.go"}, // same repo, different unit
	}

	if err := g.mustUpdate(ctx, t, "a/b", "a/b/u", "t", testRefs1); err != nil {
		t.Fatal(err)
	}
	if err := g.mustUpdate(ctx, t, "a/b", "a/b/p", "t", testRefs2); err != nil {
		t.Fatal(err)
	}
	if err := g.mustUpdate(ctx, t, "x/y", "x/y/c", "t", testRefs3); err != nil {
		t.Fatal(err)
	}
	// Updates should be idempotent.
	if err := g.mustUpdate(ctx, t, "a/b", "a/b/p", "t", testRefs2); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		Query  sourcegraph.DefSpec
		Result []*sourcegraph.DefRepoRef
	}{
		{
			sourcegraph.DefSpec{Repo: "a/b", Unit: "a/b/u", UnitType: "t", Path: "A/R"},
			[]*sourcegraph.DefRepoRef{
				{Repo: "a/b", Count: 3, Files: []*sourcegraph.DefFileRef{{Path: "a/b/u/s.go", Count: 2}, {Path: "a/b/p/t.go", Count: 1}}},
			},
		},
		{
			sourcegraph.DefSpec{Repo: "x/y", Unit: "x/y/c", UnitType: "t", Path: "A/R"},
			[]*sourcegraph.DefRepoRef{
				{Repo: "a/b", Count: 1, Files: []*sourcegraph.DefFileRef{{Path: "a/b/u/s.go", Count: 1}}},
				{Repo: "x/y", Count: 1, Files: []*sourcegraph.DefFileRef{{Path: "x/y/c/v.go", Count: 1}}},
			},
		},
	}
	for _, test := range testCases {
		got, err := g.Get(ctx, &sourcegraph.DefsListRefLocationsOp{Def: test.Query})
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Errorf("got nil result from GlobalRefs.Get")
			continue
		}
		if !reflect.DeepEqual(got.RepoRefs, test.Result) {
			t.Errorf("got %+v, want %+v", got.RepoRefs, test.Result)
		}
	}
}
