package localstore

import (
	"math"
	"reflect"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	sgtest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

const (
	commitID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

type outerCase struct {
	defs    []*graph.Def
	queries []queryCase
}

type queryCase struct {
	query []string
	opt   sourcegraph.SearchOptions

	expResults []*sourcegraph.DefSearchResult
}

func TestDefs(t *testing.T) {
	// depends on srclibsupport
	t.Skip("https://github.com/sourcegraph/sourcegraph/issues/1276")
	if testing.Short() {
		t.Skip()
	}

	var (
		d_abc_xyz = graph.Def{DefKey: graph.DefKey{Repo: "a/b", CommitID: commitID, Unit: "a/b/u", UnitType: "GoPackage", Path: "abc/xyz"}, Name: "XYZ", Kind: "func"}
		d_xyz_abc = graph.Def{DefKey: graph.DefKey{Repo: "a/b", CommitID: commitID, Unit: "a/b/u", UnitType: "GoPackage", Path: "xyz/abc"}, Name: "ABC", Kind: "field"}
		d_pqr     = graph.Def{DefKey: graph.DefKey{Repo: "a/b", CommitID: commitID, Unit: "a/b/u", UnitType: "GoPackage", Path: "pqr"}, Name: "PQR", Kind: "field"}
	)

	tests := []outerCase{{
		defs: []*graph.Def{&d_xyz_abc},
		queries: []queryCase{
			{
				[]string{"abc"},
				sourcegraph.SearchOptions{},
				[]*sourcegraph.DefSearchResult{
					{Def: sourcegraph.Def{Def: d_xyz_abc}},
				},
			},
			{
				[]string{"asdf"},
				sourcegraph.SearchOptions{},
				[]*sourcegraph.DefSearchResult{},
			},
			{
				[]string{"xyz"},
				sourcegraph.SearchOptions{},
				[]*sourcegraph.DefSearchResult{
					{Def: sourcegraph.Def{Def: d_xyz_abc}},
				},
			},
		},
	}, {
		defs: []*graph.Def{&d_abc_xyz, &d_xyz_abc, &d_pqr},
		queries: []queryCase{
			{
				[]string{"abc"},
				sourcegraph.SearchOptions{},
				[]*sourcegraph.DefSearchResult{
					{Def: sourcegraph.Def{Def: d_xyz_abc}},
					{Def: sourcegraph.Def{Def: d_abc_xyz}},
				},
			},
			{
				[]string{"pqr"},
				sourcegraph.SearchOptions{},
				[]*sourcegraph.DefSearchResult{
					{Def: sourcegraph.Def{Def: d_pqr}},
				},
			},
			{
				[]string{"abc", "xyz"},
				sourcegraph.SearchOptions{},
				[]*sourcegraph.DefSearchResult{
					{Def: sourcegraph.Def{Def: d_abc_xyz}},
					{Def: sourcegraph.Def{Def: d_xyz_abc}},
				},
			},
			{
				[]string{"xyz", "abc"},
				sourcegraph.SearchOptions{},
				[]*sourcegraph.DefSearchResult{
					{Def: sourcegraph.Def{Def: d_xyz_abc}},
					{Def: sourcegraph.Def{Def: d_abc_xyz}},
				},
			},
		},
	}}

	for _, test := range tests {
		testDefs(t, test)
	}
}

func testDefs(t *testing.T, outerTest outerCase) {
	var g defs
	ctx, done := testContext()
	var mocks *mocks // FIXME
	defer done()
	TestMockRepos = nil

	var (
		repoURIs = make(map[string]struct{})
		rps      []*sourcegraph.Repo
		units_   = make(map[unit.Key]struct{})
		units    []*unit.SourceUnit
	)
	for _, def := range outerTest.defs {
		if _, seen := repoURIs[def.Repo]; !seen {
			rps = append(rps, &sourcegraph.Repo{URI: def.Repo})
			repoURIs[def.Repo] = struct{}{}
		}

		ukey := unit.Key{Name: def.Unit, Type: def.UnitType}
		if _, seen := units_[ukey]; !seen {
			units = append(units, &unit.SourceUnit{Key: ukey})
			units_[ukey] = struct{}{}
		}
	}

	rps = (&repos{}).mustCreate(ctx, t, rps...)

	GraphMockDefs(&mocks.Graph, outerTest.defs...)
	GraphMockUnits(&mocks.Graph, units...)
	mocks.Repos.GetByURI = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		return &sourcegraph.Repo{}, nil
	}
	mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return sgtest.MockRepository{
			ResolveRevision_: func(ctx context.Context, spec string) (vcs.CommitID, error) {
				return vcs.CommitID(commitID), nil
			},
			Branches_: func(context.Context, vcs.BranchesOptions) ([]*vcs.Branch, error) {
				return []*vcs.Branch{
					&vcs.Branch{
						Commit: &vcs.Commit{ID: vcs.CommitID(commitID)},
					},
				}, nil
			},
		}, nil
	}

	for _, repo := range rps {
		op := RefreshIndexOp{Repo: repo.ID, CommitID: commitID}
		err := g.Update(ctx, op)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, test := range outerTest.queries {
		got, err := g.Search(ctx, DefSearchOp{Opt: &test.opt, TokQuery: test.query})
		if err != nil {
			t.Fatal(err)
		}

		if got == nil {
			t.Errorf("got nil result from Defs.Search")
			continue
		}

		// strip score
		gotDefResultsNoScore := make([]*sourcegraph.DefSearchResult, len(got.DefResults))
		for i, r := range got.DefResults {
			r_ := *r
			r_.Score = 0
			gotDefResultsNoScore[i] = &r_
		}
		if !verifyResultsMatch(gotDefResultsNoScore, test.expResults) {
			t.Errorf("for query %+v, got %+v, want %+v", test.query, got.DefResults, test.expResults)
		}
	}
}

func verifyResultsMatch(got, want []*sourcegraph.DefSearchResult) bool {
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
