package backend

import (
	"fmt"
	"reflect"
	"testing"

	"context"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

func TestReposService_Get(t *testing.T) {
	var s repos
	ctx := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:  1,
		URI: "github.com/u/r",
	}
	ghrepo := &sourcegraph.Repo{
		URI: "github.com/u/r",
	}

	github.MockGetRepo_Return(ghrepo)

	calledGet := localstore.Mocks.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := localstore.Mocks.Repos.MockUpdate(t, 1)

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	// Should not be called because mock GitHub has same data as mock DB.
	if *calledUpdate {
		t.Error("calledUpdate")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_Get_UpdateMeta(t *testing.T) {
	var s repos
	ctx := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:      1,
		URI:     "github.com/u/r",
		Private: true,
	}

	github.MockGetRepo_Return(&sourcegraph.Repo{
		Description: "This is a repository",
		Private:     true,
	})

	calledGet := localstore.Mocks.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := localstore.Mocks.Repos.MockUpdate(t, 1)

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !*calledUpdate {
		t.Error("!calledUpdate")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_Get_UnauthedUpdateMeta(t *testing.T) {
	var s repos
	ctx := testContext()

	// Remove auth from testContext
	ctx = authpkg.WithActor(ctx, &authpkg.Actor{})
	ctx = accesscontrol.WithInsecureSkip(ctx, false)

	wantRepo := &sourcegraph.Repo{
		ID:  1,
		URI: "github.com/u/r",
	}

	github.MockGetRepo_Return(&sourcegraph.Repo{
		Description: "This is a repository",
	})

	calledGet := localstore.Mocks.Repos.MockGet_Return(t, wantRepo)
	var calledUpdate bool
	localstore.Mocks.Repos.Update = func(ctx context.Context, op localstore.RepoUpdate) error {
		if !accesscontrol.Skip(ctx) {
			return legacyerr.Errorf(legacyerr.PermissionDenied, "permission denied")
		}
		calledUpdate = true
		if op.ReposUpdateOp.Repo != wantRepo.ID {
			t.Errorf("got repo %q, want %q", op.ReposUpdateOp.Repo, wantRepo.ID)
			return legacyerr.Errorf(legacyerr.NotFound, "repo %v not found", wantRepo.ID)
		}
		return nil
	}

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !calledUpdate {
		t.Error("!calledUpdate")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_Get_NonGitHub(t *testing.T) {
	var s repos
	ctx := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:  1,
		URI: "r",
	}

	github.MockGetRepo_Return(&sourcegraph.Repo{})

	calledGet := localstore.Mocks.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := localstore.Mocks.Repos.MockUpdate(t, 1)

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if *calledUpdate {
		t.Error("calledUpdate")
	}
	if !reflect.DeepEqual(repo, wantRepo) {
		t.Errorf("got %+v, want %+v", repo, wantRepo)
	}
}

func TestReposService_List(t *testing.T) {
	var s repos
	ctx := testContext()

	wantRepos := &sourcegraph.RepoList{
		Repos: []*sourcegraph.Repo{
			{URI: "r1"},
			{URI: "r2"},
		},
	}

	calledList := localstore.Mocks.Repos.MockList(t, "r1", "r2")

	repos, err := s.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !*calledList {
		t.Error("!calledList")
	}
	if !reflect.DeepEqual(repos, wantRepos) {
		t.Errorf("got %+v, want %+v", repos, wantRepos)
	}
}

func TestRepos_List_remoteOnly(t *testing.T) {
	var s repos
	ctx := testContext()

	calledListAccessible := github.MockListAccessibleRepos_Return([]*sourcegraph.Repo{
		&sourcegraph.Repo{URI: "github.com/is/accessible"},
	})
	calledReposStoreList := localstore.Mocks.Repos.MockList(t, "a/b", "github.com/is/accessible", "github.com/not/accessible")
	ctx = authpkg.WithActor(ctx, &authpkg.Actor{UID: "1", Login: "test", GitHubToken: "test"})

	repoList, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "github.com/is/accessible"}}}
	if !reflect.DeepEqual(repoList, want) {
		t.Fatalf("got repos %v, want %v", repoList, want)
	}
	if !*calledListAccessible {
		t.Error("!calledListAccessible")
	}
	if *calledReposStoreList {
		t.Error("calledReposStoreList (should not hit the repos store if RemoteOnly is true)")
	}
}

func TestRepos_List_remoteSearch(t *testing.T) {
	var s repos
	ctx := testContext()

	{
		testcase := "auth'd user (common case)"
		calledGHSearch := github.MockSearch_Return([]*sourcegraph.Repo{{URI: "remote1"}})
		calledReposList := localstore.Mocks.Repos.MockList(t, "local1")

		want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "local1"}, {URI: "remote1"}}}
		results, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteSearch: true, Query: "my query"})
		if err != nil {
			t.Fatal(err)
		}

		if !*calledGHSearch {
			t.Errorf("in test case %q, !calledGHSearch", testcase)
		}
		if !*calledReposList {
			t.Errorf("in test case %q, !calledReposList", testcase)
		}
		if !reflect.DeepEqual(want, results) {
			t.Errorf("in test case %q, wanted %+v, but got %+v", testcase, want, results)
		}
	}

	{
		testcase := "no auth'd user"
		ctx := authpkg.WithoutActor(ctx) // unauth'd context
		calledGHSearch := github.MockSearch_Return([]*sourcegraph.Repo{{URI: "remote1"}})
		calledReposList := localstore.Mocks.Repos.MockList(t, "local1")

		want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "local1"}}}
		results, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteSearch: true, Query: "my query"})
		if err != nil {
			t.Fatal(err)
		}

		if *calledGHSearch {
			t.Errorf("in test case %q, calledGHSearch", testcase)
		}
		if !*calledReposList {
			t.Errorf("in test case %q, !calledReposList", testcase)
		}
		if !reflect.DeepEqual(want, results) {
			t.Errorf("in test case %q, wanted %+v, but got %+v", testcase, want, results)
		}
	}

	{
		testcase := "no dupe when GitHub and Sourcegraph repos overlap"
		calledReposList := localstore.Mocks.Repos.MockList(t, "r1", "r2")
		calledGHSearch := github.MockSearch_Return([]*sourcegraph.Repo{{URI: "r2"}, {URI: "r3"}})

		want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "r1"}, {URI: "r2"}, {URI: "r3"}}}
		results, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteSearch: true, Query: "my query"})
		if err != nil {
			t.Fatal(err)
		}

		if !*calledGHSearch {
			t.Errorf("in test case %q, !calledGHSearch", testcase)
		}
		if !*calledReposList {
			t.Errorf("in test case %q, !calledReposList", testcase)
		}
		if !reflect.DeepEqual(want, results) {
			t.Errorf("in test case %q, wanted %+v, but got %+v", testcase, want, results)
		}
	}

	{
		testcase := "GitHub API error, e.g., search rate limit"
		calledGHSearch := false
		github.SearchRepoMock = func(ctx context.Context, query string, op *gogithub.SearchOptions) ([]*sourcegraph.Repo, error) {
			calledGHSearch = true
			return nil, fmt.Errorf("GH API error")
		}
		calledReposList := localstore.Mocks.Repos.MockList(t, "local1")

		want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "local1"}}}
		results, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteSearch: true, Query: "my query"})
		if err != nil {
			t.Fatal(err)
		}

		if !calledGHSearch {
			t.Errorf("in test case %q, !calledGHSearch", testcase)
		}
		if !*calledReposList {
			t.Errorf("in test case %q, !calledReposList", testcase)
		}
		if !reflect.DeepEqual(want, results) {
			t.Errorf("in test case %q, wanted %+v, but got %+v", testcase, want, results)
		}
	}

	{
		testcase := "github query reformatting (gorilla/mux -> user:gorilla in:name mux)"
		var lastReceviedGithubQuery string
		github.SearchRepoMock = func(ctx context.Context, query string, op *gogithub.SearchOptions) ([]*sourcegraph.Repo, error) {
			lastReceviedGithubQuery = query
			return nil, nil
		}
		localstore.Mocks.Repos.MockList(t)

		queryPairs := [][2]string{
			[2]string{"gorilla/mux", "user:gorilla in:name mux"},
			[2]string{"gorilla mux", "user:gorilla in:name mux"},
			[2]string{"mux", "mux"},
			[2]string{"gorilla mux http", "gorilla mux http"},
		}
		for _, queryPair := range queryPairs {
			_, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteSearch: true, Query: queryPair[0]})
			if err != nil {
				t.Fatal(err)
			}
			if lastReceviedGithubQuery != queryPair[1] {
				t.Errorf("in test case %q, with input query %q, wanted GitHub query %q, but got %q", testcase, queryPair[0], queryPair[1], lastReceviedGithubQuery)
			}
		}
	}
}
