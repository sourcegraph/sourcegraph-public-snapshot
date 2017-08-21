package backend

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"context"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
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

	repo, err := s.Get(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	// Should not be called because mock GitHub has same data as mock DB.
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

func TestRepos_List_remoteOnly_appInstalled(t *testing.T) {
	var s repos
	ctx := testContext()

	calledListAccessible := github.MockListAccessibleRepos_Return([]*sourcegraph.Repo{
		&sourcegraph.Repo{URI: "github.com/is/accessible-private"},
		&sourcegraph.Repo{URI: "github.com/is/accessible-public"},
	})
	calledListPublic := github.MockListPublicRepos_Return([]*sourcegraph.Repo{})
	calledReposStoreList := localstore.Mocks.Repos.MockList(t, "a/b", "github.com/is/accessible-private", "github.com/not/accessible")
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "1", Login: "test", GitHubToken: "test"})

	repoList, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "github.com/is/accessible-private"}, {URI: "github.com/is/accessible-public"}}}
	if !reflect.DeepEqual(repoList, want) {
		t.Fatalf("got repos %v, want %v", repoList, want)
	}
	if !*calledListAccessible {
		t.Error("!calledListAccessible")
	}
	if *calledListPublic {
		t.Error("calledListPublic (should not call ListPublicRepos if RemoteOnly is true and app is installed, since ListAccessibleRepos includes public repos)")
	}
	if *calledReposStoreList {
		t.Error("calledReposStoreList (should not hit the repos store if RemoteOnly is true)")
	}
}

func TestRepos_List_remoteOnly_appNotInstalled(t *testing.T) {
	var s repos
	ctx := testContext()

	calledListAccessible := github.MockListAccessibleRepos_Return([]*sourcegraph.Repo{})
	calledListPublic := github.MockListPublicRepos_Return([]*sourcegraph.Repo{{URI: "github.com/is/accessible-public"}})
	calledReposStoreList := localstore.Mocks.Repos.MockList(t, "a/b", "github.com/is/accessible-public", "github.com/not/accessible")
	ctx = actor.WithActor(ctx, &actor.Actor{UID: "1", Login: "test", GitHubToken: "test"})

	repoList, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "github.com/is/accessible-public"}}}
	if !reflect.DeepEqual(repoList, want) {
		t.Fatalf("got repos %v, want %v", repoList, want)
	}
	if !*calledListAccessible {
		t.Error("!calledListAccessible")
	}
	if !*calledListPublic {
		t.Error("!calledListPublic (should be called when calledListAccessible returns no accessible repos")
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

		github.MockListPublicRepos_Return([]*sourcegraph.Repo{{URI: "local1"}})
		github.MockSearch_Return([]*sourcegraph.Repo{{URI: "remote1"}})
		calledReposList := localstore.Mocks.Repos.MockList(t, "local1")

		want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "local1"}, {URI: "remote1"}}}
		results, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteSearch: true, Query: "my query"})
		if err != nil {
			t.Fatal(err)
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
		ctx := actor.WithoutActor(ctx) // unauth'd context
		calledGHSearch := github.MockSearch_Return([]*sourcegraph.Repo{{URI: "remote1"}})
		calledReposList := localstore.Mocks.Repos.MockList(t, "local1")

		_, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteSearch: true, Query: "my query"})
		if want := "refusing to perform remote search"; err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("in test case %q, got error %q, want non-nil containing %q", testcase, err, want)
		}

		if *calledGHSearch {
			t.Errorf("in test case %q, calledGHSearch", testcase)
		}
		if !*calledReposList {
			t.Errorf("in test case %q, !calledReposList", testcase)
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
