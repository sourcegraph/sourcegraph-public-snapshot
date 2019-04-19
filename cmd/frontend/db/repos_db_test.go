package db

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

/*
 * Helpers
 */

func sortedRepoNames(repos []*types.Repo) []api.RepoName {
	names := repoNames(repos)
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names
}

func repoNames(repos []*types.Repo) []api.RepoName {
	var names []api.RepoName
	for _, repo := range repos {
		names = append(names, repo.Name)
	}
	return names
}

func createRepo(ctx context.Context, t *testing.T, repo *types.Repo) {
	if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: repo.Name, Description: repo.Description, Fork: repo.Fork, Enabled: true}); err != nil {
		t.Fatal(err)
	}
}

func mustCreate(ctx context.Context, t *testing.T, repos ...*types.Repo) []*types.Repo {
	var createdRepos []*types.Repo
	for _, repo := range repos {
		createRepo(ctx, t, repo)
		repo, err := Repos.GetByName(ctx, repo.Name)
		if err != nil {
			t.Fatal(err)
		}
		createdRepos = append(createdRepos, repo)
	}
	return createdRepos
}

/*
 * Tests
 */

func TestRepos_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	want := mustCreate(ctx, t, &types.Repo{
		Name: "r",
		ExternalRepo: &api.ExternalRepoSpec{
			ID:          "a",
			ServiceType: "b",
			ServiceID:   "c",
		},
	})

	repo, err := Repos.Get(ctx, want[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(t, repo, want[0]) {
		t.Errorf("got %v, want %v", repo, want[0])
	}
}

func TestRepos_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
		return repos, nil
	}
	defer func() { mockAuthzFilter = nil }()

	ctx := dbtesting.TestContext(t)
	ctx = actor.WithActor(ctx, &actor.Actor{})

	want := mustCreate(ctx, t, &types.Repo{Name: "r"})

	repos, err := Repos.List(ctx, ReposListOptions{Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(t, repos, want) {
		t.Errorf("got %v, want %v", repos, want)
	}
}

func TestRepos_List_fork(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
		return repos, nil
	}
	defer func() { mockAuthzFilter = nil }()
	ctx := dbtesting.TestContext(t)
	ctx = actor.WithActor(ctx, &actor.Actor{})

	mine := mustCreate(ctx, t, &types.Repo{Name: "a/r", Fork: false})
	yours := mustCreate(ctx, t, &types.Repo{Name: "b/r", Fork: true})

	{
		repos, err := Repos.List(ctx, ReposListOptions{Enabled: true, OnlyForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, yours, repos)
	}
	{
		repos, err := Repos.List(ctx, ReposListOptions{Enabled: true, NoForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, mine, repos)
	}
	{
		repos, err := Repos.List(ctx, ReposListOptions{Enabled: true, NoForks: true, OnlyForks: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, nil, repos)
	}
	{
		repos, err := Repos.List(ctx, ReposListOptions{Enabled: true})
		if err != nil {
			t.Fatal(err)
		}
		assertJSONEqual(t, append(append([]*types.Repo(nil), mine...), yours...), repos)
	}
}

func TestRepos_List_pagination(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
		return repos, nil
	}
	defer func() { mockAuthzFilter = nil }()
	ctx := dbtesting.TestContext(t)
	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{Name: "r1"},
		{Name: "r2"},
		{Name: "r3"},
	}
	for _, repo := range createdRepos {
		mustCreate(ctx, t, repo)
	}

	type testcase struct {
		limit  int
		offset int
		exp    []api.RepoName
	}
	tests := []testcase{
		{limit: 1, offset: 0, exp: []api.RepoName{"r1"}},
		{limit: 1, offset: 1, exp: []api.RepoName{"r2"}},
		{limit: 1, offset: 2, exp: []api.RepoName{"r3"}},
		{limit: 2, offset: 0, exp: []api.RepoName{"r1", "r2"}},
		{limit: 2, offset: 2, exp: []api.RepoName{"r3"}},
		{limit: 3, offset: 0, exp: []api.RepoName{"r1", "r2", "r3"}},
		{limit: 3, offset: 3, exp: nil},
		{limit: 4, offset: 0, exp: []api.RepoName{"r1", "r2", "r3"}},
		{limit: 4, offset: 4, exp: nil},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{Enabled: true, LimitOffset: &LimitOffset{Limit: test.limit, Offset: test.offset}})
		if err != nil {
			t.Fatal(err)
		}
		if got := sortedRepoNames(repos); !reflect.DeepEqual(got, test.exp) {
			t.Errorf("for test case %v, got %v (want %v)", test, got, test.exp)
		}
	}
}

// TestRepos_List_query tests the behavior of Repos.List when called with
// a query.
// Test batch 1 (correct filtering)
func TestRepos_List_query1(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
		return repos, nil
	}
	defer func() { mockAuthzFilter = nil }()
	ctx := dbtesting.TestContext(t)
	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{Name: "abc/def"},
		{Name: "def/ghi"},
		{Name: "jkl/mno/pqr"},
		{Name: "github.com/abc/xyz"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		query string
		want  []api.RepoName
	}{
		{"def", []api.RepoName{"abc/def", "def/ghi"}},
		{"ABC/DEF", []api.RepoName{"abc/def"}},
		{"xyz", []api.RepoName{"github.com/abc/xyz"}},
		{"mno/p", []api.RepoName{"jkl/mno/pqr"}},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{Query: test.query, Enabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%q: got repos %q, want %q", test.query, got, test.want)
		}
	}
}

// Test batch 2 (correct ranking)
func TestRepos_List_query2(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
		return repos, nil
	}
	defer func() { mockAuthzFilter = nil }()
	ctx := dbtesting.TestContext(t)
	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{Name: "a/def"},
		{Name: "b/def"},
		{Name: "c/def"},
		{Name: "def/ghi"},
		{Name: "def/jkl"},
		{Name: "def/mno"},
		{Name: "abc/m"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		query string
		want  []api.RepoName
	}{
		{"def", []api.RepoName{"a/def", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"}},
		{"b/def", []api.RepoName{"b/def"}},
		{"def/", []api.RepoName{"def/ghi", "def/jkl", "def/mno"}},
		{"def/m", []api.RepoName{"def/mno"}},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{Query: test.query, Enabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("Unexpected repo result for query %q:\ngot:  %q\nwant: %q", test.query, got, test.want)
		}
	}
}

// Test sort
func TestRepos_List_sort(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
		return repos, nil
	}
	defer func() { mockAuthzFilter = nil }()
	ctx := dbtesting.TestContext(t)
	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{Name: "c/def"},
		{Name: "def/mno"},
		{Name: "b/def"},
		{Name: "abc/m"},
		{Name: "abc/def"},
		{Name: "def/jkl"},
		{Name: "def/ghi"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		query   string
		orderBy RepoListOrderBy
		want    []api.RepoName
	}{
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListName,
			}},
			want: []api.RepoName{"abc/def", "abc/m", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListCreatedAt,
			}},
			want: []api.RepoName{"c/def", "def/mno", "b/def", "abc/m", "abc/def", "def/jkl", "def/ghi"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCreatedAt,
				Descending: true,
			}},
			want: []api.RepoName{"def/ghi", "def/jkl", "abc/def", "abc/m", "b/def", "def/mno", "c/def"},
		},
		{
			query: "def",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCreatedAt,
				Descending: true,
			}},
			want: []api.RepoName{"def/ghi", "def/jkl", "abc/def", "b/def", "def/mno", "c/def"},
		},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{Query: test.query, OrderBy: test.orderBy, Enabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("Unexpected repo result for query %q, orderBy %v:\ngot:  %q\nwant: %q", test.query, test.orderBy, got, test.want)
		}
	}
}

// TestRepos_List_patterns tests the behavior of Repos.List when called with
// IncludePatterns and ExcludePattern.
func TestRepos_List_patterns(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
		return repos, nil
	}
	defer func() { mockAuthzFilter = nil }()
	ctx := dbtesting.TestContext(t)
	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{Name: "a/b"},
		{Name: "c/d"},
		{Name: "e/f"},
		{Name: "g/h"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		includePatterns []string
		excludePattern  string
		want            []api.RepoName
	}{
		{
			includePatterns: []string{"(a|c)"},
			want:            []api.RepoName{"a/b", "c/d"},
		},
		{
			includePatterns: []string{"(a|c)", "b"},
			want:            []api.RepoName{"a/b"},
		},
		{
			includePatterns: []string{"(a|c)"},
			excludePattern:  "d",
			want:            []api.RepoName{"a/b"},
		},
		{
			excludePattern: "(d|e)",
			want:           []api.RepoName{"a/b", "g/h"},
		},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{
			IncludePatterns: test.includePatterns,
			ExcludePattern:  test.excludePattern,
			Enabled:         true,
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("include %q exclude %q: got repos %q, want %q", test.includePatterns, test.excludePattern, got, test.want)
		}
	}
}

// TestRepos_List_patterns tests the behavior of Repos.List when called with
// a QueryPattern.
func TestRepos_List_queryPattern(t *testing.T) {
	mockAuthzFilter = func(ctx context.Context, repos []*types.Repo, p authz.Perm) ([]*types.Repo, error) {
		return repos, nil
	}
	defer func() { mockAuthzFilter = nil }()
	ctx := dbtesting.TestContext(t)
	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{Name: "a/b"},
		{Name: "c/d"},
		{Name: "e/f"},
		{Name: "g/h"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		q    query.Q
		want []api.RepoName
		err  string
	}{
		// These are the same tests as TestRepos_List_patterns, but in an
		// expression form.
		{
			q:    "(a|c)",
			want: []api.RepoName{"a/b", "c/d"},
		},
		{
			q:    query.And("(a|c)", "b"),
			want: []api.RepoName{"a/b"},
		},
		{
			q:    query.And("(a|c)", query.Not("d")),
			want: []api.RepoName{"a/b"},
		},
		{
			q:    query.Not("(d|e)"),
			want: []api.RepoName{"a/b", "g/h"},
		},

		// Some extra tests which test the pattern compiler
		{
			q:    "",
			want: []api.RepoName{"a/b", "c/d", "e/f", "g/h"},
		},
		{
			q:    "^a/b$",
			want: []api.RepoName{"a/b"},
		},
		{
			// Should match only e/f, but pattern compiler doesn't handle this
			// so matches nothing.
			q:    "[a-zA-Z]/e",
			want: nil,
		},

		// Test OR support
		{
			q:    query.Or(query.Not("(d|e)"), "d"),
			want: []api.RepoName{"a/b", "c/d", "g/h"},
		},

		// Test deeply nested
		{
			q: query.Or(
				query.And(
					true,
					query.Not(query.Or("a", "c"))),
				query.And(query.Not("e"), query.Not("a"))),
			want: []api.RepoName{"c/d", "e/f", "g/h"},
		},

		// Corner cases for Or
		{
			q:    query.Or(), // empty Or is false
			want: nil,
		},
		{
			q:    query.Or("a"),
			want: []api.RepoName{"a/b"},
		},

		// Corner cases for And
		{
			q:    query.And(), // empty And is true
			want: []api.RepoName{"a/b", "c/d", "e/f", "g/h"},
		},
		{
			q:    query.And("a"),
			want: []api.RepoName{"a/b"},
		},
		{
			q:    query.And("a", "d"),
			want: nil,
		},

		// Bad pattern
		{
			q:   query.And("a/b", ")*"),
			err: "error parsing regexp",
		},
		// Only want strings
		{
			q:   query.And("a/b", 1),
			err: "unexpected token",
		},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{
			PatternQuery: test.q,
			Enabled:      true,
		})
		if err != nil {
			if test.err == "" {
				t.Fatal(err)
			}
			if !strings.Contains(err.Error(), test.err) {
				t.Errorf("expected error to contain %q, got: %v", test.err, err)
			}
			continue
		}
		if test.err != "" {
			t.Errorf("%s: expected error", query.Print(test.q))
			continue
		}
		if got := repoNames(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: got repos %q, want %q", query.Print(test.q), got, test.want)
		}
	}
}

func TestRepos_List_queryAndPatternsMutuallyExclusive(t *testing.T) {
	ctx := context.Background()
	wantErr := "Query and IncludePatterns/ExcludePattern options are mutually exclusive"

	t.Run("Query and IncludePatterns", func(t *testing.T) {
		_, err := Repos.List(ctx, ReposListOptions{Query: "x", IncludePatterns: []string{"y"}, Enabled: true})
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("got error %v, want it to contain %q", err, wantErr)
		}
	})

	t.Run("Query and ExcludePattern", func(t *testing.T) {
		_, err := Repos.List(ctx, ReposListOptions{Query: "x", ExcludePattern: "y", Enabled: true})
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("got error %v, want it to contain %q", err, wantErr)
		}
	})
}

func TestRepos_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	// Add a repo.
	createRepo(ctx, t, &types.Repo{Name: "a/b"})

	repo, err := Repos.GetByName(ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if repo.CreatedAt.IsZero() {
		t.Fatal("got CreatedAt.IsZero()")
	}
}

func TestRepos_Create_dupe(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	// Add a repo.
	createRepo(ctx, t, &types.Repo{Name: "a/b"})

	// Add another repo with the same name.
	createRepo(ctx, t, &types.Repo{Name: "a/b"})
}
