package backend

import (
	"reflect"
	"testing"

	githubpkg "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

func TestReposService_Get(t *testing.T) {
	var s repos
	ctx := testContext()

	wantRepo := &types.Repo{
		ID:      1,
		URI:     "github.com/u/r",
		Enabled: true,
	}

	github.MockGetRepo_Return(&githubpkg.Repository{FullName: strptr("u/r")})

	calledGet := db.Mocks.Repos.MockGet_Return(t, wantRepo)

	repo, err := s.Get(ctx, &types.RepoSpec{ID: 1})
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

	wantRepos := []*types.Repo{
		{URI: "r1"},
		{URI: "r2"},
	}

	calledList := db.Mocks.Repos.MockList(t, "r1", "r2")

	repos, err := s.List(ctx, db.ReposListOptions{Enabled: true})
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

func strptr(s string) *string { return &s }
