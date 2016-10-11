package localstore

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"gopkg.in/gorp.v1"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	sgtest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/srclib/graph"
	sstore "sourcegraph.com/sourcegraph/srclib/store"
)

func TestGlobalRefs(t *testing.T) {
	t.Skip("https://github.com/sourcegraph/sourcegraph/issues/1276")
	if testing.Short() {
		t.Skip()
	}

	g := &deprecatedGlobalRefs{}

	ctx, done := testContext()
	defer done()

	createdRepos := (&repos{}).mustCreate(ctx, t, &sourcegraph.Repo{URI: "x/y"}, &sourcegraph.Repo{URI: "a/b"})
	xyRepoID := createdRepos[0].ID
	abRepoID := createdRepos[1].ID

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
	mockRefs(allRefs)
	for repo := range allRefs {
		repoObj, err := (&repos{}).GetByURI(ctx, repo)
		if err != nil {
			t.Fatal(err)
		}
		if err := g.DeprecatedUpdate(ctx, DeprecatedRefreshIndexOp{Repo: repoObj.ID, CommitID: "aaaaa"}); err != nil {
			t.Fatalf("could not update %s: %s", repo, err)
		}
	}
	// Updates should be idempotent.
	err := g.DeprecatedUpdate(ctx, DeprecatedRefreshIndexOp{Repo: abRepoID, CommitID: "aaaaa"})
	if err != nil {
		t.Fatalf("could not idempotent update a/b: %s", err)
	}

	testCases := map[string]struct {
		Op     *sourcegraph.DeprecatedDefsListRefLocationsOp
		Result []*sourcegraph.DeprecatedDefRepoRef
	}{
		"simple1": {
			&sourcegraph.DeprecatedDefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: abRepoID, Unit: "a/b/u", UnitType: "t", Path: "A/R"},
			},
			[]*sourcegraph.DeprecatedDefRepoRef{
				{Repo: "a/b", Count: 3, Files: []*sourcegraph.DeprecatedDefFileRef{{Path: "a/b/u/s.go", Count: 2}, {Path: "a/b/p/t.go", Count: 1}}},
			},
		},
		"simple2": {
			&sourcegraph.DeprecatedDefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: xyRepoID, Unit: "x/y/c", UnitType: "t", Path: "A/R"},
			},
			[]*sourcegraph.DeprecatedDefRepoRef{
				{Repo: "x/y", Count: 1, Files: []*sourcegraph.DeprecatedDefFileRef{{Path: "x/y/c/v.go", Count: 1}}},
				{Repo: "a/b", Count: 1, Files: []*sourcegraph.DeprecatedDefFileRef{{Path: "a/b/u/s.go", Count: 1}}},
			},
		},
		"repo": {
			&sourcegraph.DeprecatedDefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: xyRepoID, Unit: "x/y/c", UnitType: "t", Path: "A/R"},
				Opt: &sourcegraph.DeprecatedDefListRefLocationsOptions{
					Repos: []string{"a/b"},
				},
			},
			[]*sourcegraph.DeprecatedDefRepoRef{
				{Repo: "a/b", Count: 1, Files: []*sourcegraph.DeprecatedDefFileRef{{Path: "a/b/u/s.go", Count: 1}}},
			},
		},
		"pagination_first": {
			&sourcegraph.DeprecatedDefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: xyRepoID, Unit: "x/y/c", UnitType: "t", Path: "A/R"},
				Opt: &sourcegraph.DeprecatedDefListRefLocationsOptions{
					ListOptions: sourcegraph.ListOptions{
						Page: 1,
					},
				},
			},
			[]*sourcegraph.DeprecatedDefRepoRef{
				{Repo: "x/y", Count: 1, Files: []*sourcegraph.DeprecatedDefFileRef{{Path: "x/y/c/v.go", Count: 1}}},
				{Repo: "a/b", Count: 1, Files: []*sourcegraph.DeprecatedDefFileRef{{Path: "a/b/u/s.go", Count: 1}}},
			},
		},
		"pagination_empty": {
			&sourcegraph.DeprecatedDefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: xyRepoID, Unit: "x/y/c", UnitType: "t", Path: "A/R"},
				Opt: &sourcegraph.DeprecatedDefListRefLocationsOptions{
					ListOptions: sourcegraph.ListOptions{
						Page: 100,
					},
				},
			},
			[]*sourcegraph.DeprecatedDefRepoRef{},
		},
		// Missing defspec should not return an error
		"empty": {
			&sourcegraph.DeprecatedDefsListRefLocationsOp{
				Def: sourcegraph.DefSpec{Repo: xyRepoID, Unit: "x/y/c", UnitType: "t", Path: "A/R/D"},
			},
			[]*sourcegraph.DeprecatedDefRepoRef{},
		},
	}
	for tn, test := range testCases {
		if tn != "repo" {
			continue
		}
		got, err := g.DeprecatedGet(ctx, test.Op)
		if err != nil {
			t.Fatal(err)
		}
		if got == nil {
			t.Errorf("%s: got nil result from DeprecatedGlobalRefs.Get", tn)
			continue
		}
		if !reflect.DeepEqual(got.RepoRefs, test.Result) {
			t.Errorf("%s: got %+v, want %+v", tn, got.RepoRefs, test.Result)
		}
	}
}

func TestGlobalRefsUpdate(t *testing.T) {
	t.Skip("https://github.com/sourcegraph/sourcegraph/issues/1276")
	if testing.Short() {
		t.Skip()
	}

	g := &deprecatedGlobalRefs{}
	ctx, done := testContext()
	defer done()

	createdRepos := (&repos{}).mustCreate(ctx, t, &sourcegraph.Repo{URI: "def/repo"}, &sourcegraph.Repo{URI: "repo"})
	defRepoID := createdRepos[0].ID
	repoID := createdRepos[1].ID

	allRefs := map[string][]*graph.Ref{}
	mockRefs(allRefs)

	def := sourcegraph.DefSpec{Repo: defRepoID, Unit: "def/unit", UnitType: "def/type", Path: "def/path"}
	nFiles := 10
	genRefs := func(dir string) {
		refs := make([]*graph.Ref, 0, nFiles)
		for i := 0; i < nFiles; i++ {
			refs = append(refs, &graph.Ref{
				DefRepo:     "def/repo",
				DefUnit:     def.Unit,
				DefUnitType: def.UnitType,
				DefPath:     def.Path,
				File:        fmt.Sprintf("%s/file%d.go", dir, i),
				Repo:        "repo",
				Unit:        "unit",
				UnitType:    "unitType",
			})
		}
		allRefs["repo"] = refs
	}

	query := &sourcegraph.DeprecatedDefsListRefLocationsOp{Def: def}
	check := func(tn, dir string) {
		got, err := g.DeprecatedGet(ctx, query)
		if err != nil {
			t.Fatalf("%s: %s", tn, err)
		}
		if len(got.RepoRefs) != 1 {
			t.Fatalf("%s: expected only 1 repo, got %d", tn, len(got.RepoRefs))
		}
		expected := make([]*sourcegraph.DeprecatedDefFileRef, 0, nFiles)
		for i := 0; i < nFiles; i++ {
			expected = append(expected, &sourcegraph.DeprecatedDefFileRef{
				Path:  fmt.Sprintf("%s/file%d.go", dir, i),
				Count: 1,
			})
		}
		if !reflect.DeepEqual(got.RepoRefs[0].Files, expected) {
			t.Fatalf("%s: got unexpected DeprecatedDefFileRefs. got=%v expected=%v", tn, got.RepoRefs[0].Files, expected)
		}
	}

	// Initially we should have no refs
	got, err := g.DeprecatedGet(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.RepoRefs) != 0 {
		t.Fatalf("Expected only %d refs, got %d", 0, len(got.RepoRefs))
	}

	// We should only have results for first
	genRefs("first")
	err = g.DeprecatedUpdate(ctx, DeprecatedRefreshIndexOp{Repo: repoID, CommitID: "aaaaa"})
	if err != nil {
		t.Fatal(err)
	}
	check("first", "first")

	// We always reindex, even if we have indexed a commit.
	genRefs("second")
	err = g.DeprecatedUpdate(ctx, DeprecatedRefreshIndexOp{Repo: repoID, CommitID: "aaaaa"})
	if err != nil {
		t.Fatal(err)
	}
	check("second", "second")

	// Update what the latest commit is, that should cause us to index third
	genRefs("third")
	err = g.DeprecatedUpdate(ctx, DeprecatedRefreshIndexOp{Repo: repoID, CommitID: "bbbbb"})
	if err != nil {
		t.Fatal(err)
	}
	check("third", "third")
}

// TestGlobalRefs_version checks that we are getting the locking semantics we
// want on the global_refs_version table
func TestGlobalRefs_version(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	g := &deprecatedGlobalRefs{}
	ctx, done := testContext()
	defer done()

	get := func(want string) {
		got, err := g.version(graphDBH(ctx), "r")
		if err != nil {
			t.Fatalf("Failed to get when expecting %s", want)
		}
		if got != want {
			t.Fatalf("version is %+v, wanted %+v", got, want)
		}
	}
	update := func(tx gorp.SqlExecutor, commitID string) {
		err := g.versionUpdate(tx, "r", commitID)
		if err != nil {
			t.Fatalf("Failed to update to %s", commitID)
		}
	}

	// nothing set yet
	get("")
	// just put something in
	update(graphDBH(ctx), "first")

	// now do a get and an update in the background. The update should
	// happen first and the transaction for it lasts longer. So we expect
	// get to happen afterwards
	wg := sync.WaitGroup{}
	wg.Add(2)
	var txFinished, getFinished time.Time
	var err error
	go func() {
		defer wg.Done()
		err = dbutil.Transact(graphDBH(ctx), func(tx gorp.SqlExecutor) error {
			update(tx, "tx")
			time.Sleep(100 * time.Millisecond)
			txFinished = time.Now()
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		get("tx")
		getFinished = time.Now()
	}()
	wg.Wait()
	if getFinished.Before(txFinished) {
		t.Fatalf("concurrent get finished %s before transaction.", txFinished.Sub(getFinished))
	}
}

func benchmarkGlobalRefsGet(b *testing.B) {
	ctx, done := testContext()
	defer done()
	get := func() error {
		repo, err := (&repos{}).GetByURI(ctx, "github.com/golang/go")
		if err != nil {
			return err
		}
		_, err = DeprecatedGlobalRefs.DeprecatedGet(ctx, &sourcegraph.DeprecatedDefsListRefLocationsOp{Def: sourcegraph.DefSpec{Repo: repo.ID, Unit: "fmt", UnitType: "GoPackage", Path: "Errorf"}})
		return err
	}
	if err := get(); err != nil {
		b.Log("Loading data into GlobalRefs")
		nRepos := 10000
		nRefs := 10
		deprecatedGlobalRefsUpdate(b, ctx, nRepos, nRefs)
		b.Log("Refreshing")
		DeprecatedGlobalRefs.DeprecatedStatRefresh(ctx)
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

func benchmarkGlobalRefsUpdate(b *testing.B) {
	ctx, done := testContext()
	defer done()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		deprecatedGlobalRefsUpdate(b, ctx, 1, 100)
	}
	// defer done() can be expensive
	b.StopTimer()
}

func deprecatedGlobalRefsUpdate(b *testing.B, ctx context.Context, nRepos, nRefs int) {
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
	mockRefs(allRefs)
	for i := 0; i < nRepos; i++ {
		pkg := fmt.Sprintf("foo.com/foo/bar%d", i)
		repoObj, err := (&repos{}).GetByURI(ctx, pkg)
		if err != nil {
			b.Fatal(err)
		}
		if err := DeprecatedGlobalRefs.DeprecatedUpdate(ctx, DeprecatedRefreshIndexOp{Repo: repoObj.ID, CommitID: "aaaaa"}); err != nil {
			b.Fatal(err)
		}
	}
}

func mockRefs(allRefs map[string][]*graph.Ref) {
	Mocks.Graph.Refs_ = func(f ...sstore.RefFilter) ([]*graph.Ref, error) {
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
	Mocks.Repos.Get = func(ctx context.Context, repo int32) (*sourcegraph.Repo, error) {
		return &sourcegraph.Repo{}, nil
	}
	Mocks.RepoVCS.Open = func(ctx context.Context, repo int32) (vcs.Repository, error) {
		return sgtest.MockRepository{
			ResolveRevision_: func(ctx context.Context, spec string) (vcs.CommitID, error) {
				return "aaaa", nil
			},
		}, nil
	}
}
