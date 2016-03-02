package handlerutil

import (
	"net/url"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
)

func TestRepoFallback(t *testing.T) {
	config.RepoFallbackURLStr = "fallback-server"
	if err := config.parseURL(); err != nil {
		t.Fatal(err)
	}

	sourcegraph.MockNewClientFromContext(func(ctx context.Context) (*sourcegraph.Client, error) {
		switch sourcegraph.GRPCEndpoint(ctx).String() {
		case "fallback-server":
			return &sourcegraph.Client{
				Repos: &mock.ReposClient{
					Get_: func(context.Context, *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
						return &sourcegraph.Repo{URI: "myrepo"}, nil
					},
				},
			}, nil

		case "main-server":
			return &sourcegraph.Client{
				Repos: &mock.ReposClient{
					Get_: func(context.Context, *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
						return nil, grpc.Errorf(codes.NotFound, "")
					},
				},
			}, nil

		default:
			panic("unhandled endpoint")
		}
	})
	defer sourcegraph.RestoreNewClientFromContext()

	ctx := context.Background()
	ctx = sourcegraph.WithGRPCEndpoint(ctx, &url.URL{Path: "main-server"})

	repo, _, err := GetRepo(ctx, sourcegraph.RepoSpec{URI: "myrepo"}.RouteVars())
	if err != nil {
		t.Fatal(err)
	}
	if want := "myrepo"; repo.URI != want {
		t.Errorf("got %q, want %q", repo.URI, want)
	}
}
