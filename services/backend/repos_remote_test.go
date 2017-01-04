package backend

import (
	"reflect"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	githubmock "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/mocks"
)

func TestRepos_Resolve_local(t *testing.T) {
	ctx := testContext()

	calledReposGet := localstore.Mocks.Repos.MockGetByURI(t, "r", 1)

	res, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "r"})
	if err != nil {
		t.Fatal(err)
	}
	if !*calledReposGet {
		t.Error("!calledReposGet")
	}

	want := &sourcegraph.RepoResolution{Repo: 1, CanonicalPath: "r"}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got %#v, want %#v", res, want)
	}
}

func TestRepos_Resolve_local_otherError(t *testing.T) {
	ctx := testContext()
	var githubRepos githubmock.GitHubRepoGetter
	ctx = github.WithRepos(ctx, &githubRepos)

	var calledReposGet bool
	localstore.Mocks.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, legacyerr.Errorf(legacyerr.Internal, "")
	}

	var calledGetGitHubRepo bool
	githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return nil, legacyerr.Errorf(legacyerr.Internal, "")
	}

	_, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "r"})
	if legacyerr.ErrCode(err) != legacyerr.Internal {
		t.Errorf("got error %v, want Internal", err)
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
	if calledGetGitHubRepo {
		t.Error("calledGetGitHubRepo (should only be called after Repos.Get returns NotFound)")
	}
}

func TestRepos_Resolve_GitHub_NonRemote(t *testing.T) {
	ctx := testContext()
	var githubRepos githubmock.GitHubRepoGetter
	ctx = github.WithRepos(ctx, &githubRepos)

	var calledReposGet bool
	localstore.Mocks.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, legacyerr.Errorf(legacyerr.NotFound, "")
	}

	var calledGetGitHubRepo bool
	githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return &sourcegraph.Repo{Name: "github.com/o/r"}, nil
	}

	if _, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "github.com/o/r", Remote: false}); legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Errorf("got error %v, want NotFound", err)
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
	if !calledGetGitHubRepo {
		t.Error("!calledGetGitHubRepo")
	}
}

func TestRepos_Resolve_GitHub_Remote(t *testing.T) {
	ctx := testContext()
	var githubRepos githubmock.GitHubRepoGetter
	ctx = github.WithRepos(ctx, &githubRepos)

	var calledReposGet bool
	localstore.Mocks.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, legacyerr.Errorf(legacyerr.NotFound, "")
	}

	var calledGetGitHubRepo bool
	githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return &sourcegraph.Repo{Name: "github.com/o/r"}, nil
	}

	res, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "github.com/o/r", Remote: true})
	if err != nil {
		t.Fatal(err)
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
	if !calledGetGitHubRepo {
		t.Error("!calledGetGitHubRepo")
	}

	want := &sourcegraph.RepoResolution{RemoteRepo: &sourcegraph.Repo{Name: "github.com/o/r"}}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got %#v, want %#v", res, want)
	}
}

func TestRepos_Resolve_GitHub_otherError(t *testing.T) {
	ctx := testContext()
	var githubRepos githubmock.GitHubRepoGetter
	ctx = github.WithRepos(ctx, &githubRepos)

	var calledReposGet bool
	localstore.Mocks.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, legacyerr.Errorf(legacyerr.NotFound, "")
	}

	var calledGetGitHubRepo bool
	githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return nil, legacyerr.Errorf(legacyerr.Internal, "")
	}

	_, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "github.com/o/r"})
	if legacyerr.ErrCode(err) != legacyerr.Internal {
		t.Errorf("got error %v, want Internal", err)
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
	if !calledGetGitHubRepo {
		t.Error("!calledGetGitHubRepo")
	}
}

func TestRepos_Resolve_notFound(t *testing.T) {
	ctx := testContext()
	var githubRepos githubmock.GitHubRepoGetter
	ctx = github.WithRepos(ctx, &githubRepos)

	var calledReposGet bool
	localstore.Mocks.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, legacyerr.Errorf(legacyerr.NotFound, "")
	}

	var calledGetGitHubRepo bool
	githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return nil, legacyerr.Errorf(legacyerr.NotFound, "")
	}

	_, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "github.com/o/r"})
	if legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Errorf("got error %v, want NotFound", err)
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
	if !calledGetGitHubRepo {
		t.Error("!calledGetGitHubRepo")
	}
}

func TestRepos_Resolve_other_notFound(t *testing.T) {
	ctx := testContext()
	var githubRepos githubmock.GitHubRepoGetter
	ctx = github.WithRepos(ctx, &githubRepos)

	var calledReposGet bool
	localstore.Mocks.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, legacyerr.Errorf(legacyerr.NotFound, "")
	}

	var calledGetGitHubRepo bool
	githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return nil, legacyerr.Errorf(legacyerr.NotFound, "")
	}

	_, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "r"})
	if legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Errorf("got error %v, want NotFound", err)
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
	if calledGetGitHubRepo {
		t.Error("githubRepos.Get was called, but shouldn't be, since repo URI doesn't have 'github.com/' prefix")
	}
}
