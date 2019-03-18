package db

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"context"

	dbtesting "github.com/sourcegraph/sourcegraph/cmd/frontend/db/testing"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

/*
 * Helpers
 */

func sortedRepoURIs(repos []*types.Repo) []api.RepoURI {
	uris := repoURIs(repos)
	sort.Slice(uris, func(i, j int) bool { return uris[i] < uris[j] })
	return uris
}

func repoURIs(repos []*types.Repo) []api.RepoURI {
	var uris []api.RepoURI
	for _, repo := range repos {
		uris = append(uris, repo.URI)
	}
	return uris
}

func createRepo(ctx context.Context, t *testing.T, repo *types.Repo) {
	if err := Repos.Upsert(ctx, api.InsertRepoOp{URI: repo.URI, Description: repo.Description, Fork: repo.Fork, Enabled: true}); err != nil {
		t.Fatal(err)
	}
}

func mustCreate(ctx context.Context, t *testing.T, repos ...*types.Repo) []*types.Repo {
	var createdRepos []*types.Repo
	for _, repo := range repos {
		createRepo(ctx, t, repo)
		repo, err := Repos.GetByURI(ctx, repo.URI)
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
		URI: "r",
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

	ctx := dbtesting.TestContext(t)

	ctx = actor.WithActor(ctx, &actor.Actor{})

	want := mustCreate(ctx, t, &types.Repo{URI: "r"})

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

	ctx := dbtesting.TestContext(t)

	ctx = actor.WithActor(ctx, &actor.Actor{})

	mine := mustCreate(ctx, t, &types.Repo{URI: "a/r", Fork: false})
	yours := mustCreate(ctx, t, &types.Repo{URI: "b/r", Fork: true})

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

	ctx := dbtesting.TestContext(t)

	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{URI: "r1"},
		{URI: "r2"},
		{URI: "r3"},
	}
	for _, repo := range createdRepos {
		mustCreate(ctx, t, repo)
	}

	type testcase struct {
		limit  int
		offset int
		exp    []api.RepoURI
	}
	tests := []testcase{
		{limit: 1, offset: 0, exp: []api.RepoURI{"r1"}},
		{limit: 1, offset: 1, exp: []api.RepoURI{"r2"}},
		{limit: 1, offset: 2, exp: []api.RepoURI{"r3"}},
		{limit: 2, offset: 0, exp: []api.RepoURI{"r1", "r2"}},
		{limit: 2, offset: 2, exp: []api.RepoURI{"r3"}},
		{limit: 3, offset: 0, exp: []api.RepoURI{"r1", "r2", "r3"}},
		{limit: 3, offset: 3, exp: nil},
		{limit: 4, offset: 0, exp: []api.RepoURI{"r1", "r2", "r3"}},
		{limit: 4, offset: 4, exp: nil},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{Enabled: true, LimitOffset: &LimitOffset{Limit: test.limit, Offset: test.offset}})
		if err != nil {
			t.Fatal(err)
		}
		if got := sortedRepoURIs(repos); !reflect.DeepEqual(got, test.exp) {
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

	ctx := dbtesting.TestContext(t)

	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{URI: "abc/def"},
		{URI: "def/ghi"},
		{URI: "jkl/mno/pqr"},
		{URI: "github.com/abc/xyz"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		query string
		want  []api.RepoURI
	}{
		{"def", []api.RepoURI{"abc/def", "def/ghi"}},
		{"ABC/DEF", []api.RepoURI{"abc/def"}},
		{"xyz", []api.RepoURI{"github.com/abc/xyz"}},
		{"mno/p", []api.RepoURI{"jkl/mno/pqr"}},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{Query: test.query, Enabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoURIs(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%q: got repos %q, want %q", test.query, got, test.want)
		}
	}
}

// Test batch 2 (correct ranking)
func TestRepos_List_query2(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{URI: "a/def"},
		{URI: "b/def"},
		{URI: "c/def"},
		{URI: "def/ghi"},
		{URI: "def/jkl"},
		{URI: "def/mno"},
		{URI: "abc/m"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		query string
		want  []api.RepoURI
	}{
		{"def", []api.RepoURI{"a/def", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"}},
		{"b/def", []api.RepoURI{"b/def"}},
		{"def/", []api.RepoURI{"def/ghi", "def/jkl", "def/mno"}},
		{"def/m", []api.RepoURI{"def/mno"}},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{Query: test.query, Enabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoURIs(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("Unexpected repo result for query %q:\ngot:  %q\nwant: %q", test.query, got, test.want)
		}
	}
}

// Test indexed_revision
func TestRepos_List_indexedRevision(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{URI: "a/def", IndexedRevision: (*api.CommitID)(strptr("aaaaaa"))},
		{URI: "b/def"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
		gotRepo, err := Repos.GetByURI(ctx, repo.URI)
		if err != nil {
			panic(err)
		}
		if repo.IndexedRevision != nil {
			Repos.UpdateIndexedRevision(ctx, gotRepo.ID, *repo.IndexedRevision)
		}
	}
	tests := []struct {
		hasIndexedRevision *bool
		want               []api.RepoURI
	}{
		{nil, []api.RepoURI{"a/def", "b/def"}},
		{boolptr(true), []api.RepoURI{"a/def"}},
		{boolptr(false), []api.RepoURI{"b/def"}},
	}

	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{HasIndexedRevision: test.hasIndexedRevision, Enabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoURIs(repos); !reflect.DeepEqual(got, test.want) {
			if test.hasIndexedRevision == nil {
				t.Errorf("Unexpected repo result for hasIndexedRevision %v:\ngot:  %q\nwant: %q", test.hasIndexedRevision, got, test.want)
			} else {
				t.Errorf("Unexpected repo result for hasIndexedRevision %v:\ngot:  %q\nwant: %q", *test.hasIndexedRevision, got, test.want)
			}
		}
	}
}

// Test sort
func TestRepos_List_sort(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{URI: "c/def"},
		{URI: "def/mno"},
		{URI: "b/def"},
		{URI: "abc/m"},
		{URI: "abc/def"},
		{URI: "def/jkl"},
		{URI: "def/ghi"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		query   string
		orderBy RepoListOrderBy
		want    []api.RepoURI
	}{
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListURI,
			}},
			want: []api.RepoURI{"abc/def", "abc/m", "b/def", "c/def", "def/ghi", "def/jkl", "def/mno"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field: RepoListCreatedAt,
			}},
			want: []api.RepoURI{"c/def", "def/mno", "b/def", "abc/m", "abc/def", "def/jkl", "def/ghi"},
		},
		{
			query: "",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCreatedAt,
				Descending: true,
			}},
			want: []api.RepoURI{"def/ghi", "def/jkl", "abc/def", "abc/m", "b/def", "def/mno", "c/def"},
		},
		{
			query: "def",
			orderBy: RepoListOrderBy{{
				Field:      RepoListCreatedAt,
				Descending: true,
			}},
			want: []api.RepoURI{"def/ghi", "def/jkl", "abc/def", "b/def", "def/mno", "c/def"},
		},
	}
	for _, test := range tests {
		repos, err := Repos.List(ctx, ReposListOptions{Query: test.query, OrderBy: test.orderBy, Enabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoURIs(repos); !reflect.DeepEqual(got, test.want) {
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

	ctx := dbtesting.TestContext(t)

	ctx = actor.WithActor(ctx, &actor.Actor{})

	createdRepos := []*types.Repo{
		{URI: "a/b"},
		{URI: "c/d"},
		{URI: "e/f"},
		{URI: "g/h"},
	}
	for _, repo := range createdRepos {
		createRepo(ctx, t, repo)
	}
	tests := []struct {
		includePatterns []string
		excludePattern  string
		want            []api.RepoURI
	}{
		{
			includePatterns: []string{"(a|c)"},
			want:            []api.RepoURI{"a/b", "c/d"},
		},
		{
			includePatterns: []string{"(a|c)", "b"},
			want:            []api.RepoURI{"a/b"},
		},
		{
			includePatterns: []string{"(a|c)"},
			excludePattern:  "d",
			want:            []api.RepoURI{"a/b"},
		},
		{
			excludePattern: "(d|e)",
			want:           []api.RepoURI{"a/b", "g/h"},
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
		if got := repoURIs(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("include %q exclude %q: got repos %q, want %q", test.includePatterns, test.excludePattern, got, test.want)
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
	createRepo(ctx, t, &types.Repo{URI: "a/b"})

	repo, err := Repos.GetByURI(ctx, "a/b")
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
	createRepo(ctx, t, &types.Repo{URI: "a/b"})

	// Add another repo with the same name.
	createRepo(ctx, t, &types.Repo{URI: "a/b"})
}

func boolptr(b bool) *bool {
	return &b
}
