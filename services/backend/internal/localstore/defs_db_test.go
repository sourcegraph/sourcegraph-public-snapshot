package localstore

import (
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store/mockstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	sgtest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func TestDefs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	var g defs
	ctx, mocks, done := testContext()
	defer done()
	ctx = store.WithRepos(ctx, &repos{})

	repos := (&repos{}).mustCreate(ctx, t, &sourcegraph.Repo{URI: "a/b"})
	commitID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	testDefs1 := []*graph.Def{
		{DefKey: graph.DefKey{Repo: "a/b", CommitID: commitID, Unit: "a/b/u", UnitType: "GoPackage", Path: "abc/xyz"}, Name: "XYZ", Kind: "func", File: "a.go"},
		{DefKey: graph.DefKey{Repo: "a/b", CommitID: commitID, Unit: "a/b/u", UnitType: "GoPackage", Path: "xyz/abc"}, Name: "ABC", Kind: "field", File: "a.go"},
		{DefKey: graph.DefKey{Repo: "a/b", CommitID: commitID, Unit: "a/b/u", UnitType: "GoPackage", Path: "pqr"}, Name: "PQR", Kind: "field", File: "b.go"},
	}

	mockstore.GraphMockDefs(&mocks.Stores.Graph, testDefs1...)
	mockstore.GraphMockUnits(&mocks.Stores.Graph, &unit.SourceUnit{Key: unit.Key{Name: "a/b/u", Type: "t"}})
	mocks.Repos.GetByURI_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		return &sourcegraph.Repo{}, nil
	}
	mocks.RepoVCS.Open_ = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return sgtest.MockRepository{
			ResolveRevision_: func(spec string) (vcs.CommitID, error) {
				return vcs.CommitID(commitID), nil
			},
			Branches_: func(vcs.BranchesOptions) ([]*vcs.Branch, error) {
				return []*vcs.Branch{
					&vcs.Branch{
						Commit: &vcs.Commit{ID: vcs.CommitID(commitID)},
					},
				}, nil
			},
		}, nil
	}

	op := store.DefUpdateOp{Repo: repos[0].ID, CommitID: commitID}
	err := g.UpdateFromSrclibStore(ctx, op)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		Query   []string
		Results []*sourcegraph.DefSearchResult
	}{
		{
			[]string{"abc"},
			[]*sourcegraph.DefSearchResult{
				{Def: sourcegraph.Def{Def: *testDefs1[1]}},
			},
		},
		{
			[]string{"pqr"},
			[]*sourcegraph.DefSearchResult{
				{Def: sourcegraph.Def{Def: *testDefs1[2]}},
			},
		},
		{
			[]string{"abc", "xyz"},
			[]*sourcegraph.DefSearchResult{
				{Def: sourcegraph.Def{Def: *testDefs1[0]}},
				{Def: sourcegraph.Def{Def: *testDefs1[1]}},
			},
		},
		{
			[]string{"xyz", "abc"},
			[]*sourcegraph.DefSearchResult{
				{Def: sourcegraph.Def{Def: *testDefs1[1]}},
				{Def: sourcegraph.Def{Def: *testDefs1[0]}},
			},
		},
	}
	for _, test := range testCases {
		got, err := g.Search(ctx, store.DefSearchOp{Opt: &sourcegraph.SearchOptions{Repos: []int32{repos[0].ID}, CommitID: commitID}, TokQuery: test.Query})
		if err != nil {
			t.Fatal(err)
		}

		if got == nil {
			t.Errorf("got nil result from GlobalDefs.Search")
			continue
		}

		// strip score
		for _, res := range got.DefResults {
			res.Score = 0
		}

		if !verifyResultsMatch(got.DefResults, test.Results) {
			t.Errorf("got %+v, want %+v", got.DefResults, test.Results)
		}
	}
}
