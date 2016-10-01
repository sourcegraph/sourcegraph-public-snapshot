package backend

import (
	"reflect"
	"testing"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestRepos_Resolve_local(t *testing.T) {
	ctx, mock := testContext()

	calledReposGet := mock.stores.Repos.MockGetByURI(t, "r", 1)

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
	ctx, mock := testContext()

	var calledReposGet bool
	mock.stores.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, grpc.Errorf(codes.Internal, "")
	}

	var calledGetGitHubRepo bool
	mock.githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return nil, grpc.Errorf(codes.Internal, "")
	}

	_, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "r"})
	if grpc.Code(err) != codes.Internal {
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
	ctx, mock := testContext()

	var calledReposGet bool
	mock.stores.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	var calledGetGitHubRepo bool
	mock.githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return &sourcegraph.Repo{Origin: &sourcegraph.Origin{ID: "123", Service: sourcegraph.Origin_GitHub}}, nil
	}

	if _, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "r", Remote: false}); grpc.Code(err) != codes.NotFound {
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
	ctx, mock := testContext()

	var calledReposGet bool
	mock.stores.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	var calledGetGitHubRepo bool
	mock.githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return &sourcegraph.Repo{Origin: &sourcegraph.Origin{ID: "123", Service: sourcegraph.Origin_GitHub}}, nil
	}

	res, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "r", Remote: true})
	if err != nil {
		t.Fatal(err)
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
	if !calledGetGitHubRepo {
		t.Error("!calledGetGitHubRepo")
	}

	want := &sourcegraph.RepoResolution{RemoteRepo: &sourcegraph.Repo{Origin: &sourcegraph.Origin{ID: "123", Service: sourcegraph.Origin_GitHub}}}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got %#v, want %#v", res, want)
	}
}

func TestRepos_Resolve_GitHub_otherError(t *testing.T) {
	ctx, mock := testContext()

	var calledReposGet bool
	mock.stores.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	var calledGetGitHubRepo bool
	mock.githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return nil, grpc.Errorf(codes.Internal, "")
	}

	_, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "r"})
	if grpc.Code(err) != codes.Internal {
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
	ctx, mock := testContext()

	var calledReposGet bool
	mock.stores.Repos.GetByURI = func(context.Context, string) (*sourcegraph.Repo, error) {
		calledReposGet = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	var calledGetGitHubRepo bool
	mock.githubRepos.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		calledGetGitHubRepo = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	_, err := (&repos{}).Resolve(ctx, &sourcegraph.RepoResolveOp{Path: "r"})
	if grpc.Code(err) != codes.NotFound {
		t.Errorf("got error %v, want NotFound", err)
	}
	if !calledReposGet {
		t.Error("!calledReposGet")
	}
	if !calledGetGitHubRepo {
		t.Error("!calledGetGitHubRepo")
	}
}
