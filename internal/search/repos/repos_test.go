package repos

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestRevisionValidation(t *testing.T) {
	// mocks a repo repoFoo with revisions revBar and revBas
	gitserver.Mocks.ResolveRevision = func(spec string, opt gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		// trigger errors
		if spec == "bad_commit" {
			return "", gitdomain.BadCommitError{}
		}
		if spec == "deadline_exceeded" {
			return "", context.DeadlineExceeded
		}

		// known revisions
		m := map[string]struct{}{
			"revBar": {},
			"revBas": {},
		}
		if _, ok := m[spec]; ok {
			return "", nil
		}
		return "", &gitdomain.RevisionNotFoundError{Repo: "repoFoo", Spec: spec}
	}
	defer func() { gitserver.Mocks.ResolveRevision = nil }()

	tests := []struct {
		repoFilters              []string
		wantRepoRevs             []*search.RepositoryRevisions
		wantMissingRepoRevisions []*search.RepositoryRevisions
		wantErr                  error
	}{
		{
			repoFilters: []string{"repoFoo@revBar:^revBas"},
			wantRepoRevs: []*search.RepositoryRevisions{{
				Repo: types.MinimalRepo{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "revBar",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
					{
						RevSpec:        "^revBas",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
				},
			}},
			wantMissingRepoRevisions: nil,
		},
		{
			repoFilters: []string{"repoFoo@*revBar:*!revBas"},
			wantRepoRevs: []*search.RepositoryRevisions{{
				Repo: types.MinimalRepo{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "",
						RefGlob:        "revBar",
						ExcludeRefGlob: "",
					},
					{
						RevSpec:        "",
						RefGlob:        "",
						ExcludeRefGlob: "revBas",
					},
				},
			}},
			wantMissingRepoRevisions: nil,
		},
		{
			repoFilters: []string{"repoFoo@revBar:^revQux"},
			wantRepoRevs: []*search.RepositoryRevisions{{
				Repo: types.MinimalRepo{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "revBar",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
				},
				ListRefs: nil,
			}},
			wantMissingRepoRevisions: []*search.RepositoryRevisions{{
				Repo: types.MinimalRepo{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "^revQux",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
				},
			}},
			wantErr: &MissingRepoRevsError{},
		},
		{
			repoFilters:              []string{"repoFoo@revBar:bad_commit"},
			wantRepoRevs:             nil,
			wantMissingRepoRevisions: nil,
			wantErr:                  gitdomain.BadCommitError{},
		},
		{
			repoFilters:              []string{"repoFoo@revBar:^bad_commit"},
			wantRepoRevs:             nil,
			wantMissingRepoRevisions: nil,
			wantErr:                  gitdomain.BadCommitError{},
		},
		{
			repoFilters:              []string{"repoFoo@revBar:deadline_exceeded"},
			wantRepoRevs:             nil,
			wantMissingRepoRevisions: nil,
			wantErr:                  context.DeadlineExceeded,
		},
		{
			repoFilters: []string{"repoFoo"},
			wantRepoRevs: []*search.RepositoryRevisions{{
				Repo: types.MinimalRepo{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
				},
			}},
			wantMissingRepoRevisions: nil,
			wantErr:                  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.repoFilters[0], func(t *testing.T) {
			repos := database.NewMockRepoStore()
			repos.ListMinimalReposFunc.SetDefaultReturn([]types.MinimalRepo{{Name: "repoFoo"}}, nil)
			db := database.NewMockDB()
			db.ReposFunc.SetDefaultReturn(repos)

			op := search.RepoOptions{RepoFilters: tt.repoFilters}
			repositoryResolver := &Resolver{DB: db}
			resolved, err := repositoryResolver.Resolve(context.Background(), op)

			if diff := cmp.Diff(tt.wantRepoRevs, resolved.RepoRevs); diff != "" {
				t.Error(diff)
			}
			if diff := cmp.Diff(tt.wantMissingRepoRevisions, resolved.MissingRepoRevs); diff != "" {
				t.Error(diff)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got: %v, expected: %v", err, tt.wantErr)
			}
			mockrequire.Called(t, repos.ListMinimalReposFunc)
		})
	}
}

// TestSearchRevspecs tests a repository name against a list of
// repository specs with optional revspecs, and determines whether
// we get the expected error, list of matching rev specs, or list
// of clashing revspecs (if no matching rev specs were found)
func TestSearchRevspecs(t *testing.T) {
	type testCase struct {
		descr    string
		specs    []string
		repo     string
		err      error
		matched  []search.RevisionSpecifier
		clashing []search.RevisionSpecifier
	}

	tests := []testCase{
		{
			descr:    "simple match",
			specs:    []string{"foo"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: ""}},
			clashing: nil,
		},
		{
			descr:    "single revspec",
			specs:    []string{".*o@123456"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		{
			descr:    "revspec plus unspecified rev",
			specs:    []string{".*o@123456", "foo"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		{
			descr:    "revspec plus unspecified rev, but backwards",
			specs:    []string{".*o", "foo@123456"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "123456"}},
			clashing: nil,
		},
		{
			descr:    "conflicting revspecs",
			specs:    []string{".*o@123456", "foo@234567"},
			repo:     "foo",
			err:      nil,
			matched:  nil,
			clashing: []search.RevisionSpecifier{{RevSpec: "123456"}, {RevSpec: "234567"}},
		},
		{
			descr:    "overlapping revspecs",
			specs:    []string{".*o@a:b", "foo@b:c"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "b"}},
			clashing: nil,
		},
		{
			descr:    "multiple overlapping revspecs",
			specs:    []string{".*o@a:b:c", "foo@b:c:d"},
			repo:     "foo",
			err:      nil,
			matched:  []search.RevisionSpecifier{{RevSpec: "b"}, {RevSpec: "c"}},
			clashing: nil,
		},
		{
			descr:    "invalid regexp",
			specs:    []string{"*o@a:b"},
			repo:     "foo",
			err:      errors.Errorf("%s", "bad request: in findPatternRevs: error parsing regexp: missing argument to repetition operator: `*`"),
			matched:  nil,
			clashing: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.descr, func(t *testing.T) {
			_, pats, err := findPatternRevs(test.specs)
			if err != nil {
				if test.err == nil {
					t.Errorf("unexpected error: '%s'", err)
				}
				if test.err != nil && err.Error() != test.err.Error() {
					t.Errorf("incorrect error: got '%s', expected '%s'", err, test.err)
				}
				// don't try to use the pattern list if we got an error
				return
			}
			if test.err != nil {
				t.Errorf("missing expected error: wanted '%s'", test.err.Error())
			}
			matched, clashing := getRevsForMatchedRepo(api.RepoName(test.repo), pats)
			if !reflect.DeepEqual(matched, test.matched) {
				t.Errorf("matched repo mismatch: actual: %#v, expected: %#v", matched, test.matched)
			}
			if !reflect.DeepEqual(clashing, test.clashing) {
				t.Errorf("clashing repo mismatch: actual: %#v, expected: %#v", clashing, test.clashing)
			}
		})
	}
}

func BenchmarkGetRevsForMatchedRepo(b *testing.B) {
	b.Run("2 conflicting", func(b *testing.B) {
		_, pats, _ := findPatternRevs([]string{".*o@123456", "foo@234567"})
		for i := 0; i < b.N; i++ {
			_, _ = getRevsForMatchedRepo("foo", pats)
		}
	})

	b.Run("multiple overlapping", func(b *testing.B) {
		_, pats, _ := findPatternRevs([]string{".*o@a:b:c:d", "foo@b:c:d:e", "foo@c:d:e:f"})
		for i := 0; i < b.N; i++ {
			_, _ = getRevsForMatchedRepo("foo", pats)
		}
	})
}

func TestResolverPaginate(t *testing.T) {
	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))

	for i := 1; i <= 5; i++ {
		r := types.MinimalRepo{
			Name:  api.RepoName(fmt.Sprintf("github.com/foo/bar%d", i)),
			Stars: i * 100,
		}

		if err := db.Repos().Create(ctx, r.ToRepo()); err != nil {
			t.Fatal(err)
		}
	}

	all, err := (&Resolver{DB: db}).Resolve(ctx, search.RepoOptions{})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name  string
		opts  search.RepoOptions
		pages []Resolved
		err   error
	}{
		{
			name:  "default limit 500, no cursors",
			opts:  search.RepoOptions{},
			pages: []Resolved{all},
		},
		{
			name: "with limit 3, no cursors",
			opts: search.RepoOptions{
				Limit: 3,
			},
			pages: []Resolved{
				{
					RepoRevs: all.RepoRevs[:3],
					Next: types.MultiCursor{
						{Column: "stars", Direction: "prev", Value: fmt.Sprint(all.RepoRevs[3].Repo.Stars)},
						{Column: "id", Direction: "prev", Value: fmt.Sprint(all.RepoRevs[3].Repo.ID)},
					},
				},
				{
					RepoRevs: all.RepoRevs[3:],
				},
			},
		},
		{
			name: "with limit 3 and cursor",
			opts: search.RepoOptions{
				Limit: 3,
				Cursors: types.MultiCursor{
					{Column: "stars", Direction: "prev", Value: fmt.Sprint(all.RepoRevs[3].Repo.Stars)},
					{Column: "id", Direction: "prev", Value: fmt.Sprint(all.RepoRevs[3].Repo.ID)},
				},
			},
			pages: []Resolved{
				{
					RepoRevs: all.RepoRevs[3:],
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := Resolver{Opts: tc.opts, DB: db}

			var pages []Resolved
			err := r.Paginate(ctx, func(page *Resolved) error {
				pages = append(pages, *page)
				return nil
			})
			if err != nil {
				t.Error(err)
			}

			if !errors.Is(err, tc.err) {
				t.Errorf("%s unexpected error (-have, +want):\n%s", tc.name, cmp.Diff(err, tc.err))
			}

			if diff := cmp.Diff(pages, tc.pages); diff != "" {
				t.Errorf("%s unexpected pages (-have, +want):\n%s", tc.name, diff)
			}
		})
	}
}

func TestResolveRepositoriesWithUserSearchContext(t *testing.T) {
	const (
		wantName   = "alice"
		wantUserID = 123
	)

	repos := database.NewMockRepoStore()
	repos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, op database.ReposListOptions) ([]types.MinimalRepo, error) {
		if op.UserID != wantUserID {
			t.Fatalf("got %q, want %q", op.UserID, wantUserID)
		}
		return []types.MinimalRepo{
			{
				ID:   1,
				Name: "example.com/a",
			},
			{
				ID:   2,
				Name: "example.com/b",
			},
			{
				ID:   3,
				Name: "example.com/c",
			},
			{
				ID:   4,
				Name: "external.com/a",
			},
			{
				ID:   5,
				Name: "external.com/b",
			},
			{
				ID:   6,
				Name: "external.com/c",
			},
		}, nil
	})

	ns := database.NewMockNamespaceStore()
	ns.GetByNameFunc.SetDefaultHook(func(ctx context.Context, name string) (*database.Namespace, error) {
		if name != wantName {
			t.Fatalf("got %q, want %q", name, wantName)
		}
		return &database.Namespace{Name: wantName, User: wantUserID}, nil
	})

	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)
	db.NamespacesFunc.SetDefaultReturn(ns)

	op := search.RepoOptions{
		SearchContextSpec: "@" + wantName,
	}
	repositoryResolver := &Resolver{DB: db}
	resolved, err := repositoryResolver.Resolve(context.Background(), op)
	if err != nil {
		t.Fatal(err)
	}
	var got []api.RepoName
	for _, rev := range resolved.RepoRevs {
		got = append(got, rev.Repo.Name)
	}
	sort.Slice(got, func(i, j int) bool {
		return got[i] < got[j]
	})
	want := []api.RepoName{
		"example.com/a",
		"example.com/b",
		"example.com/c",
		"external.com/a",
		"external.com/b",
		"external.com/c",
	}
	if diff := cmp.Diff(got, want, nil); diff != "" {
		t.Errorf("unexpected diff: %s", diff)
	}

	mockrequire.Called(t, ns.GetByNameFunc)
	mockrequire.Called(t, repos.ListMinimalReposFunc)
}

func stringSliceToRevisionSpecifiers(revisions []string) []search.RevisionSpecifier {
	revisionSpecs := make([]search.RevisionSpecifier, 0, len(revisions))
	for _, revision := range revisions {
		revisionSpecs = append(revisionSpecs, search.RevisionSpecifier{RevSpec: revision})
	}
	return revisionSpecs
}

func TestResolveRepositoriesWithSearchContext(t *testing.T) {
	searchContext := &types.SearchContext{ID: 1, Name: "searchcontext"}
	repoA := types.MinimalRepo{ID: 1, Name: "example.com/a"}
	repoB := types.MinimalRepo{ID: 2, Name: "example.com/b"}
	searchContextRepositoryRevisions := []*types.SearchContextRepositoryRevisions{
		{Repo: repoA, Revisions: []string{"branch-1", "branch-3"}},
		{Repo: repoB, Revisions: []string{"branch-2"}},
	}

	gitserver.Mocks.ResolveRevision = func(spec string, opt gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		return api.CommitID(spec), nil
	}

	repos := database.NewMockRepoStore()
	repos.ListMinimalReposFunc.SetDefaultHook(func(ctx context.Context, op database.ReposListOptions) ([]types.MinimalRepo, error) {
		if op.SearchContextID != searchContext.ID {
			t.Fatalf("got %q, want %q", op.SearchContextID, searchContext.ID)
		}
		return []types.MinimalRepo{repoA, repoB}, nil
	})

	sc := database.NewMockSearchContextsStore()
	sc.GetSearchContextFunc.SetDefaultHook(func(ctx context.Context, opts database.GetSearchContextOptions) (*types.SearchContext, error) {
		if opts.Name != searchContext.Name {
			t.Fatalf("got %q, want %q", opts.Name, searchContext.Name)
		}
		return searchContext, nil
	})
	sc.GetSearchContextRepositoryRevisionsFunc.SetDefaultHook(func(ctx context.Context, searchContextID int64) ([]*types.SearchContextRepositoryRevisions, error) {
		if searchContextID != searchContext.ID {
			t.Fatalf("got %q, want %q", searchContextID, searchContext.ID)
		}
		return searchContextRepositoryRevisions, nil
	})

	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)
	db.SearchContextsFunc.SetDefaultReturn(sc)

	op := search.RepoOptions{
		SearchContextSpec: "searchcontext",
	}
	repositoryResolver := &Resolver{DB: db}
	resolved, err := repositoryResolver.Resolve(context.Background(), op)
	if err != nil {
		t.Fatal(err)
	}
	wantRepositoryRevisions := []*search.RepositoryRevisions{
		{Repo: repoA, Revs: stringSliceToRevisionSpecifiers(searchContextRepositoryRevisions[0].Revisions)},
		{Repo: repoB, Revs: stringSliceToRevisionSpecifiers(searchContextRepositoryRevisions[1].Revisions)},
	}
	if !reflect.DeepEqual(resolved.RepoRevs, wantRepositoryRevisions) {
		t.Errorf("got repository revisions %+v, want %+v", resolved.RepoRevs, wantRepositoryRevisions)
	}
}
