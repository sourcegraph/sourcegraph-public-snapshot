package backend

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

func TestReposService_Get(t *testing.T) {
	var s repos
	ctx := testContext()

	wantRepo := &types.Repo{
		ID:      1,
		URI:     "github.com/u/r",
		Enabled: true,
	}

	calledGet := db.Mocks.Repos.MockGet_Return(t, wantRepo)

	repo, err := s.Get(ctx, 1)
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

func TestRepos_Add(t *testing.T) {
	var s repos
	ctx := testContext()

	const repoURI = "my/repo"

	calledRepoLookup := false
	repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		calledRepoLookup = true
		if args.Repo != repoURI {
			t.Errorf("got %q, want %q", args.Repo, repoURI)
		}
		return &protocol.RepoLookupResult{
			Repo: &protocol.RepoInfo{URI: repoURI, Description: "d"},
		}, nil
	}
	defer func() { repoupdater.MockRepoLookup = nil }()

	calledUpsert := false
	db.Mocks.Repos.Upsert = func(op api.InsertRepoOp) error {
		calledUpsert = true
		if want := (api.InsertRepoOp{URI: repoURI, Description: "d"}); !reflect.DeepEqual(op, want) {
			t.Errorf("got %+v, want %+v", op, want)
		}
		return nil
	}

	if err := s.Add(ctx, repoURI); err != nil {
		t.Fatal(err)
	}
	if !calledRepoLookup {
		t.Error("!calledRepoLookup")
	}
	if !calledUpsert {
		t.Error("!calledUpsert")
	}
}
