package backend

import (
	"fmt"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"context"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

func TestReposService_Get(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:     1,
		URI:    "github.com/u/r",
		Mirror: true,
	}
	ghrepo := &sourcegraph.Repo{
		URI:    "github.com/u/r",
		Mirror: true,
	}

	mock.githubRepos.MockGet_Return(ctx, ghrepo)

	calledGet := mock.stores.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := mock.stores.Repos.MockUpdate(t, 1)

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
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:     1,
		URI:    "github.com/u/r",
		Mirror: true,
	}

	mock.githubRepos.MockGet_Return(ctx, &sourcegraph.Repo{
		Description: "This is a repository",
	})

	calledGet := mock.stores.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := mock.stores.Repos.MockUpdate(t, 1)

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
	ctx, mock := testContext()

	// Remove auth from testContext
	ctx = authpkg.WithActor(ctx, &authpkg.Actor{})
	ctx = accesscontrol.WithInsecureSkip(ctx, false)

	wantRepo := &sourcegraph.Repo{
		ID:     1,
		URI:    "github.com/u/r",
		Mirror: true,
	}

	mock.githubRepos.MockGet_Return(ctx, &sourcegraph.Repo{
		Description: "This is a repository",
	})

	calledGet := mock.stores.Repos.MockGet_Return(t, wantRepo)
	var calledUpdate bool
	mock.stores.Repos.Update = func(ctx context.Context, op store.RepoUpdate) error {
		if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Repos.Update", op.Repo); err != nil {
			return err
		}
		calledUpdate = true
		if op.ReposUpdateOp.Repo != wantRepo.ID {
			t.Errorf("got repo %q, want %q", op.ReposUpdateOp.Repo, wantRepo.ID)
			return grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo.ID)
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
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:          1,
		URI:         "r",
		Mirror:      true,
		Permissions: &sourcegraph.RepoPermissions{Pull: true, Push: true},
	}

	mock.githubRepos.MockGet_Return(ctx, &sourcegraph.Repo{})

	calledGet := mock.stores.Repos.MockGet_Return(t, wantRepo)
	calledUpdate := mock.stores.Repos.MockUpdate(t, 1)

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

func TestRepos_Create_New(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:   1,
		URI:  "r",
		Name: "r",
	}

	calledCreate := false
	mock.stores.Repos.Create = func(ctx context.Context, repo *sourcegraph.Repo) (int32, error) {
		calledCreate = true
		if repo.URI != wantRepo.URI {
			t.Errorf("got uri %#v, want %#v", repo.URI, wantRepo.URI)
		}
		return wantRepo.ID, nil
	}
	mock.stores.Repos.MockGet(t, 1)

	_, err := s.Create(ctx, &sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_New{New: &sourcegraph.ReposCreateOp_NewRepo{
			URI: "r",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledCreate {
		t.Error("!calledCreate")
	}
}

func TestRepos_Create_Origin(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepo := &sourcegraph.Repo{
		ID:  1,
		URI: "github.com/a/b",
		Origin: &sourcegraph.Origin{
			ID:         "123",
			Service:    sourcegraph.Origin_GitHub,
			APIBaseURL: "https://api.github.com",
		},
	}

	calledGet := false
	mock.githubRepos.GetByID_ = func(ctx context.Context, id int) (*sourcegraph.Repo, error) {
		if want := 123; id != want {
			t.Errorf("got id %d, want %d", id, want)
		}
		calledGet = true
		return &sourcegraph.Repo{Origin: &sourcegraph.Origin{ID: "123", Service: sourcegraph.Origin_GitHub}}, nil
	}

	calledCreate := false
	mock.stores.Repos.Create = func(ctx context.Context, repo *sourcegraph.Repo) (int32, error) {
		calledCreate = true
		if !reflect.DeepEqual(repo.Origin, wantRepo.Origin) {
			t.Errorf("got repo origin %#v, want %#v", repo.Origin, wantRepo.Origin)
		}
		return wantRepo.ID, nil
	}
	mock.stores.Repos.MockGet(t, 1)

	_, err := s.Create(ctx, &sourcegraph.ReposCreateOp{
		Op: &sourcegraph.ReposCreateOp_Origin{Origin: wantRepo.Origin},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledGet {
		t.Error("!calledGet")
	}
	if !calledCreate {
		t.Error("!calledCreate")
	}
}

func TestReposService_List(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepos := &sourcegraph.RepoList{
		Repos: []*sourcegraph.Repo{
			{URI: "r1"},
			{URI: "r2"},
		},
	}

	calledList := mock.stores.Repos.MockList(t, "r1", "r2")

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
	ctx, mock := testContext()

	calledListAccessible := mock.githubRepos.MockListAccessible(ctx, []*sourcegraph.Repo{
		&sourcegraph.Repo{URI: "github.com/is/accessible"},
	})
	calledReposStoreList := mock.stores.Repos.MockList(t, "a/b", "github.com/is/accessible", "github.com/not/accessible")
	ctx = github.WithMockHasAuthedUser(ctx, true)

	repoList, err := s.List(ctx, &sourcegraph.RepoListOptions{RemoteOnly: true})
	if err != nil {
		t.Fatal(err)
	}

	want := &sourcegraph.RepoList{Repos: []*sourcegraph.Repo{{URI: "github.com/is/accessible"}}}
	if !reflect.DeepEqual(repoList, want) {
		t.Fatalf("got repos %q, want %q", repoList, want)
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
	ctx, mock := testContext()

	{
		testcase := "auth'd user (common case)"
		calledGHSearch := mock.githubRepos.MockSearch(ctx, []*sourcegraph.Repo{{URI: "remote1"}})
		calledReposList := mock.stores.Repos.MockList(t, "local1")

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
		calledGHSearch := mock.githubRepos.MockSearch(ctx, []*sourcegraph.Repo{{URI: "remote1"}})
		calledReposList := mock.stores.Repos.MockList(t, "local1")

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
		calledReposList := mock.stores.Repos.MockList(t, "r1", "r2")
		calledGHSearch := mock.githubRepos.MockSearch(ctx, []*sourcegraph.Repo{{URI: "r2"}, {URI: "r3"}})

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
		mock.githubRepos.Search_ = func(ctx context.Context, query string, op *gogithub.SearchOptions) ([]*sourcegraph.Repo, error) {
			calledGHSearch = true
			return nil, fmt.Errorf("GH API error")
		}
		calledReposList := mock.stores.Repos.MockList(t, "local1")

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
		mock.githubRepos.Search_ = func(ctx context.Context, query string, op *gogithub.SearchOptions) ([]*sourcegraph.Repo, error) {
			lastReceviedGithubQuery = query
			return nil, nil
		}
		mock.stores.Repos.MockList(t)

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

func TestRepos_GetConfig(t *testing.T) {
	var s repos
	ctx, mock := testContext()

	wantRepoConfig := &sourcegraph.RepoConfig{}

	mock.stores.Repos.MockGetByURI(t, "r", 1)
	calledConfigsGet := mock.stores.RepoConfigs.MockGet_Return(t, 1, wantRepoConfig)

	conf, err := s.GetConfig(ctx, &sourcegraph.RepoSpec{ID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledConfigsGet {
		t.Error("!calledConfigsGet")
	}
	if !reflect.DeepEqual(conf, wantRepoConfig) {
		t.Errorf("got %+v, want %+v", conf, wantRepoConfig)
	}
}
