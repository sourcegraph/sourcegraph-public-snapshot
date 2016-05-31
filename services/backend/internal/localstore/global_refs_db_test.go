// +build pgsqltest

package localstore

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	sgtest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/srclib/graph"
	sstore "sourcegraph.com/sourcegraph/srclib/store"
)

func TestGlobalRefs(t *testing.T) {
	t.Parallel()
	testGlobalRefs(t, &globalRefs{})
}

func testGlobalRefs(t *testing.T, g store.GlobalRefs) {
	ctx, mocks, done := testContext()
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
		{DefPath: ".", DefRepo: "", DefUnit: "", File: "x/y/c/v.go"},        // package ref
		{DefPath: "A/R", DefRepo: "", DefUnit: "x/y/c", File: "x/y/c/v.go"}, // same unit
		{DefPath: "B/T", DefRepo: "", DefUnit: "x/y/d", File: "x/y/c/v.go"}, // same repo, different unit
	}

	allRefs := map[string][]*graph.Ref{}
	addRefs := func(repo, unitName, unitType string, refs []*graph.Ref) {
		repoRefs, ok := allRefs[repo]
		if !ok {
			repoRefs = []*graph.Ref{}
		}
		for _, rp := range refs {
			r := *rp
			if r.DefRepo == "" {
				r.DefRepo = repo
			}
			if r.DefUnit == "" {
				r.DefUnit = unitName
			}
			if r.DefUnitType == "" {
				r.DefUnitType = unitType
			}
			r.Repo = repo
			r.Unit = unitName
			r.UnitType = unitType
			repoRefs = append(repoRefs, &r)
		}
		allRefs[repo] = repoRefs
	}
	addRefs("a/b", "a/b/u", "t", testRefs1)
	addRefs("a/b", "a/b/p", "t", testRefs2)
	addRefs("x/y", "x/y/c", "t", testRefs3)
	mockRefs(mocks, allRefs)
	for repo := range allRefs {
		err := g.Update(ctx, sourcegraph.RepoSpec{URI: repo})
		if err != nil {
			t.Fatalf("could not update %s: %s", repo, err)
		}
	}
	// Updates should be idempotent.
	err := g.Update(ctx, sourcegraph.RepoSpec{URI: "a/b"})
	if err != nil {
		t.Fatalf("could not idempotent update a/b: %s", err)
	}

	testCases := map[string]struct {
		Op     *sourcegraph.DefsListRefLocationsOp
		Result []*sourcegraph.DefRepoRef
	}{
		"simple1": {
			&sourcegraph.DefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: "a/b", Unit: "a/b/u", UnitType: "t", Path: "A/R"},
			},
			[]*sourcegraph.DefRepoRef{
				{Repo: "a/b", Count: 3, Files: []*sourcegraph.DefFileRef{{Path: "a/b/u/s.go", Count: 2}, {Path: "a/b/p/t.go", Count: 1}}},
			},
		},
		"simple2": {
			&sourcegraph.DefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: "x/y", Unit: "x/y/c", UnitType: "t", Path: "A/R"},
			},
			[]*sourcegraph.DefRepoRef{
				{Repo: "x/y", Count: 1, Files: []*sourcegraph.DefFileRef{{Path: "x/y/c/v.go", Count: 1}}},
				{Repo: "a/b", Count: 1, Files: []*sourcegraph.DefFileRef{{Path: "a/b/u/s.go", Count: 1}}},
			},
		},
		"repo": {
			&sourcegraph.DefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: "x/y", Unit: "x/y/c", UnitType: "t", Path: "A/R"},
				Opt: &sourcegraph.DefListRefLocationsOptions{
					Repos: []string{"a/b"},
				},
			},
			[]*sourcegraph.DefRepoRef{
				{Repo: "a/b", Count: 1, Files: []*sourcegraph.DefFileRef{{Path: "a/b/u/s.go", Count: 1}}},
			},
		},
		"pagination_first": {
			&sourcegraph.DefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: "x/y", Unit: "x/y/c", UnitType: "t", Path: "A/R"},
				Opt: &sourcegraph.DefListRefLocationsOptions{
					ListOptions: sourcegraph.ListOptions{
						Page: 1,
					},
				},
			},
			[]*sourcegraph.DefRepoRef{
				{Repo: "x/y", Count: 1, Files: []*sourcegraph.DefFileRef{{Path: "x/y/c/v.go", Count: 1}}},
				{Repo: "a/b", Count: 1, Files: []*sourcegraph.DefFileRef{{Path: "a/b/u/s.go", Count: 1}}},
			},
		},
		"pagination_empty": {
			&sourcegraph.DefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: "x/y", Unit: "x/y/c", UnitType: "t", Path: "A/R"},
				Opt: &sourcegraph.DefListRefLocationsOptions{
					ListOptions: sourcegraph.ListOptions{
						Page: 100,
					},
				},
			},
			nil,
		},
		// Missing defspec should not return an error
		"empty": {
			&sourcegraph.DefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: "x/y", Unit: "x/y/c", UnitType: "t", Path: "A/R/D"},
			},
			nil,
		},
	}
	for tn, test := range testCases {
		got, err := g.Get(ctx, test.Op)
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Errorf("%s: got nil result from GlobalRefs.Get", tn)
			continue
		}
		if !reflect.DeepEqual(got.RepoRefs, test.Result) {
			t.Errorf("%s: got %+v, want %+v", tn, got.RepoRefs, test.Result)
		}
	}
}

func benchmarkGlobalRefsGet(b *testing.B, g store.GlobalRefs) {
	ctx, mocks, done := testContext()
	defer done()
	get := func() error {
		_, err := g.Get(ctx, &sourcegraph.DefsListRefLocationsOp{Def: sourcegraph.DefSpec{Repo: "github.com/golang/go", Unit: "fmt", UnitType: "GoPackage", Path: "Errorf"}})
		return err
	}
	if err := get(); err != nil {
		b.Log("Loading data into GlobalRefs")
		nRepos := 10000
		nRefs := 10
		globalRefsUpdate(b, g, ctx, mocks, nRepos, nRefs)
		type CanRefresh interface {
			StatRefresh(context.Context) error
		}
		if x, ok := g.(CanRefresh); ok {
			b.Log("Refreshing")
			x.StatRefresh(ctx)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := get()
		if err != nil {
			b.Fatal(err)
		}
	}

	// defer done() can be expensive
	b.StopTimer()
}

func benchmarkGlobalRefsUpdate(b *testing.B, g store.GlobalRefs) {
	ctx, mocks, done := testContext()
	defer done()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		globalRefsUpdate(b, g, ctx, mocks, 1, 100)
	}
	// defer done() can be expensive
	b.StopTimer()
}

func globalRefsUpdate(b *testing.B, g store.GlobalRefs, ctx context.Context, mocks *mocks, nRepos, nRefs int) {
	allRefs := map[string][]*graph.Ref{}
	for i := 0; i < nRepos; i++ {
		pkg := fmt.Sprintf("foo.com/foo/bar%d", i)
		refs := make([]*graph.Ref, nRefs)
		for j := 0; j < nRefs; j++ {
			file := fmt.Sprintf("foo/bar/baz%d.go", j/3)
			refs[j] = &graph.Ref{
				DefRepo:     "github.com/golang/go",
				DefUnit:     "fmt",
				DefUnitType: "GoPackage",
				DefPath:     "Errorf",
				File:        file,
				Repo:        pkg,
				UnitType:    "GoPackage",
				Unit:        pkg,
			}
		}
		repoRefs, ok := allRefs[pkg]
		if !ok {
			repoRefs = []*graph.Ref{}
		}
		repoRefs = append(repoRefs, refs...)
		allRefs[pkg] = repoRefs
	}
	mockRefs(mocks, allRefs)
	for i := 0; i < nRepos; i++ {
		pkg := fmt.Sprintf("foo.com/foo/bar%d", i)
		if err := g.Update(ctx, sourcegraph.RepoSpec{URI: pkg}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGlobalRefsGet(b *testing.B) {
	benchmarkGlobalRefsGet(b, &globalRefs{})
}

func BenchmarkGlobalRefsUpdate(b *testing.B) {
	benchmarkGlobalRefsUpdate(b, &globalRefs{})
}

func mockRefs(mocks *mocks, allRefs map[string][]*graph.Ref) {
	mocks.Stores.Graph.Refs_ = func(f ...sstore.RefFilter) ([]*graph.Ref, error) {
		if len(f) != 1 {
			return nil, errors.New("mockRefs: Expected only 1 filter")
		}
		type byRepos interface {
			ByRepos() []string
		}
		repos := f[0].(byRepos).ByRepos()
		if len(repos) != 1 {
			return nil, errors.New("mockRefs: Expected only 1 repo")
		}
		return allRefs[repos[0]], nil
	}
	mocks.Repos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		return &sourcegraph.Repo{}, nil
	}
	mocks.RepoVCS.Open_ = func(ctx context.Context, repo string) (vcs.Repository, error) {
		return sgtest.MockRepository{
			ResolveRevision_: func(spec string) (vcs.CommitID, error) {
				return "aaaa", nil
			},
		}, nil
	}
}
