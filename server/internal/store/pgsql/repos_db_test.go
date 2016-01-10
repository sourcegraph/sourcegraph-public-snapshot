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

func preCreateRepo(repo *sourcegraph.Repo) *sourcegraph.Repo {
	repo.Mirror = true
	repo.HTTPCloneURL = "http://example.com/dummy.git"
	return repo
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

func TestRepos_List_query(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_List_query(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_List_URIs(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_List_URIs(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Create(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Create(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Create_dupe(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Create_dupe(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Update_Description(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_Description(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Update_UpdatedAt(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_UpdatedAt(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Update_PushedAt(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_PushedAt(ctx, t, &repos{}, preCreateRepo)
}
