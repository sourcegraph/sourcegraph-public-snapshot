// +build pgsqltest

package pgsql

import (
	"reflect"
	"sort"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
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

	var s Repos
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

// Test the State option to Repos.List.
func TestRepos_List_stateOption(t *testing.T) {
	t.Parallel()

	r1 := &sourcegraph.Repo{URI: "r1"} // enabled because public and Enabled
	c1 := &sourcegraph.RepoConfig{Enabled: true}

	r2 := &sourcegraph.Repo{URI: "r2"} // disabled because public and Enabled==false
	c2 := &sourcegraph.RepoConfig{Enabled: false}

	r3 := &sourcegraph.Repo{URI: "r3"} // disabled because public and no conf exist
	c3 := (*sourcegraph.RepoConfig)(nil)

	var s Repos
	ctx, done := testContext()
	defer done()

	insertRepoAndConf := func(repo *sourcegraph.Repo, conf *sourcegraph.RepoConfig) {
		s.mustCreate(ctx, t, repo)
		if conf != nil {
			if err := (&RepoConfigs{}).Update(ctx, repo.URI, *conf); err != nil {
				t.Fatal(err)
			}
		}
	}
	insertRepoAndConf(r1, c1)
	insertRepoAndConf(r2, c2)
	insertRepoAndConf(r3, c3)

	getRepoURIsByState := func(state string) []string {
		repos, err := s.List(ctx, &sourcegraph.RepoListOptions{State: state})
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

	if got, want := getRepoURIsByState("enabled"), []string{"r1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("state %s: got %v, want %v", "enabled", got, want)
	}
	if got, want := getRepoURIsByState("disabled"), []string{"r2", "r3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("state %s: got %v, want %v", "disabled", got, want)
	}
	all := []string{"r1", "r2", "r3"}
	if got := getRepoURIsByState("all"); !reflect.DeepEqual(got, all) {
		t.Errorf("state %s: got %v, want %v", "all", got, all)
	}
	if got := getRepoURIsByState(""); !reflect.DeepEqual(got, all) {
		t.Errorf("state %s: got %v, want %v", "empty", got, all)
	}
}

func TestRepos_List_type(t *testing.T) {
	t.Parallel()

	r1 := &sourcegraph.Repo{URI: "r1", Private: true}
	r2 := &sourcegraph.Repo{URI: "r2"}

	var s Repos
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
	testsuite.Repos_List_query(ctx, t, &Repos{}, preCreateRepo)
}

func TestRepos_List_URIs(t *testing.T) {
	t.Parallel()
	ctx, done := testContext()
	defer done()
	testsuite.Repos_List_URIs(ctx, t, &Repos{}, preCreateRepo)
}
