package github

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/fed/discover"
	"src.sourcegraph.com/sourcegraph/svc"
)

func TestDiscoverRepoLocal_found(t *testing.T) {
	origRootFlag := fed.Config.IsRoot
	defer func() {
		fed.Config.IsRoot = origRootFlag
	}()

	fed.Config.IsRoot = true

	info, err := discover.Repo(context.Background(), "github.com/o/r")
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := info.NewContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if want := "GitHub (github.com)"; info.String() != want {
		t.Errorf("got info %q, want %q", info, want)
	}

	reposSvc := svc.Repos(ctx)
	if typ, want := reflect.TypeOf(reposSvc).String(), "*local.repos"; typ != want {
		t.Errorf("got Repos store type %q, want %q", typ, want)
	}
}

func TestDiscoverRepoLocalGHE_found(t *testing.T) {
	origRootFlag := fed.Config.IsRoot
	defer func() {
		fed.Config.IsRoot = origRootFlag
	}()

	fed.Config.IsRoot = true
	githubcli.Config.GitHubHost = "myghe.com"
	defer func() {
		githubcli.Config.GitHubHost = "github.com"
	}()

	info, err := discover.Repo(context.Background(), "myghe.com/o/r")
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := info.NewContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if want := "GitHub (myghe.com)"; info.String() != want {
		t.Errorf("got info %q, want %q", info, want)
	}

	reposSvc := svc.Repos(ctx)
	if typ, want := reflect.TypeOf(reposSvc).String(), "*local.repos"; typ != want {
		t.Errorf("got Repos store type %q, want %q", typ, want)
	}
}

func TestDiscoverRepoRemote_found(t *testing.T) {
	origRootFlag := fed.Config.IsRoot
	origRootURL := fed.Config.RootURLStr
	defer func() {
		fed.Config.IsRoot = origRootFlag
		fed.Config.RootURLStr = origRootURL
	}()

	fed.Config.IsRoot = false
	fed.Config.RootURLStr = "https://demo-mothership:13080"

	info, err := discover.Repo(context.Background(), "github.com/o/r")
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := info.NewContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if want := "GitHub (github.com)"; info.String() != want {
		t.Errorf("got info %q, want %q", info, want)
	}

	if u, want := sourcegraph.GRPCEndpoint(ctx), fed.Config.RootURLStr; u.String() != want {
		t.Errorf("got gRPC endpoint %q, want %q", u.String(), want)
	}

	reposSvc := svc.Repos(ctx)
	if typ, want := reflect.TypeOf(reposSvc).String(), "remote.remoteRepos"; typ != want {
		t.Errorf("got Repos store type %q, want %q", typ, want)
	}
}

func TestDiscover_notFound(t *testing.T) {
	// Empty out RepoFuncs to avoid falling back to HTTP discovery
	// (which hits the network and makes this test slower
	// unnecessarily).
	discover.RepoFuncs = nil

	_, err := discover.Repo(context.Background(), "example.com/foo/bar")
	if !discover.IsNotFound(err) {
		t.Fatalf("got err == %v, want *discover.NotFoundError", err)
	}
}
