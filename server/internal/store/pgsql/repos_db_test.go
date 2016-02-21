// +build pgsqltest

package pgsql

import (
	"reflect"
	"sort"
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
	"src.sourcegraph.com/sourcegraph/util/jsonutil"
)

func repoURIs(repos []*sourcegraph.Repo) []string {
	var uris []string
	for _, repo := range repos {
		uris = append(uris, repo.URI)
	}
	sort.Strings(uris)
	return uris
}

func TestRepos_List(t *testing.T) {
	t.Parallel()

	var s repos
	ctx, done := testContext()
	defer done()

	want := s.mustCreate(ctx, t, &sourcegraph.Repo{URI: "r"})

	repos, err := s.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonutil.JSONEqual(t, repos, want) {
		t.Errorf("got %v, want %v", repos, want)
	}
}

func TestRepos_List_type(t *testing.T) {
	t.Parallel()

	r1 := &sourcegraph.Repo{URI: "r1", Private: true}
	r2 := &sourcegraph.Repo{URI: "r2"}

	var s repos
	ctx, done := testContext()
	defer done()

	s.mustCreate(ctx, t, r1, r2)

	getRepoURIsByType := func(typ string) []string {
		repos, err := s.List(ctx, &sourcegraph.RepoListOptions{Type: typ})
		if err != nil {
			t.Fatal(err)
		}
		uris := make([]string, len(repos))
		for i, repo := range repos {
			uris[i] = repo.URI
		}
		sort.Strings(uris)
		return uris
	}

	if got, want := getRepoURIsByType("private"), []string{"r1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("type %s: got %v, want %v", "enabled", got, want)
	}
	if got, want := getRepoURIsByType("public"), []string{"r2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("type %s: got %v, want %v", "disabled", got, want)
	}
	all := []string{"r1", "r2"}
	if got := getRepoURIsByType("all"); !reflect.DeepEqual(got, all) {
		t.Errorf("type %s: got %v, want %v", "all", got, all)
	}
	if got := getRepoURIsByType(""); !reflect.DeepEqual(got, all) {
		t.Errorf("type %s: got %v, want %v", "empty", got, all)
	}
}

// TestRepos_List_query tests the behavior of Repos.List when called with
// a query.
func TestRepos_List_query(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()

	s := &repos{}
	// Add some repos.
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "abc/def", Name: "def", VCS: "git"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "def/ghi", Name: "ghi", VCS: "git"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, &sourcegraph.Repo{URI: "jkl/mno/pqr", Name: "pqr", VCS: "git"}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		query string
		want  []string
	}{
		{"de", []string{"abc/def", "def/ghi"}},
		{"def", []string{"abc/def", "def/ghi"}},
		{"ABC/DEF", []string{"abc/def"}},
		{"xyz", nil},
	}
	for _, test := range tests {
		repos, err := s.List(ctx, &sourcegraph.RepoListOptions{Query: test.query})
		if err != nil {
			t.Fatal(err)
		}
		if got := repoURIs(repos); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%q: got repos %v, want %v", test.query, got, test.want)
		}
	}
}

func TestRepos_List_URIs(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_List_URIs(ctx, t, &repos{})
}

func TestRepos_Create(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Create(ctx, t, &repos{})
}

func TestRepos_Create_dupe(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Create_dupe(ctx, t, &repos{})
}

func TestRepos_Update_Description(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_Description(ctx, t, &repos{})
}

func TestRepos_Update_UpdatedAt(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_UpdatedAt(ctx, t, &repos{})
}

func TestRepos_Update_PushedAt(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_PushedAt(ctx, t, &repos{})
}
