package app_test

import (
	"net/http"
	"testing"

	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/testutil/srclibtest"
)

func init() {
	graph.RegisterMakeDefFormatter("t", func(*graph.Def) graph.DefFormatter { return srclibtest.Formatter{} })
}

func TestDefVirtual(t *testing.T) {
	c, mock := apptest.New()

	wantDef := &sourcegraph.Def{
		Def: graph.Def{
			DefKey: graph.DefKey{
				Repo:     "my/repo",
				CommitID: "c",
				UnitType: "GoPackage",
				Unit:     "u",
				Path:     "__virtual__/p",
			},
		},
		DocHTML: &pbtypes.HTML{"this is a doc"},
	}

	mock.Defs.Get_ = func(ctx context.Context, in *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		return wantDef, nil
	}
	mockRepoGet(mock, "my/repo")
	mock.Repos.GetConfig_ = func(ctx context.Context, in *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error) {
		return &sourcegraph.RepoConfig{}, nil
	}
	mock.Repos.GetCommit_ = func(ctx context.Context, in *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
		return &vcs.Commit{ID: vcs.CommitID(wantDef.CommitID)}, nil
	}
	mockSpecificVersionSrclibData(mock, "c")

	resp, err := c.Get(router.Rel.URLToDef(wantDef.DefKey).String())
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected error response (%d): %s", resp.StatusCode, resp.Status)
	}

	dom, err := parseHTML(resp)
	if err != nil {
		t.Fatal(err)
	}

	// Check to see that the JavaScript contains text that shows the
	// page was displayed correctly (this is a pretty fragile test).
	js := dom.Find(".repo-body").Find("script[ignore-csp]").Contents().Text()
	if !strings.Contains(js, "This is an auto-generated file for a definition") {
		t.Errorf(`Expected auto-generated source file JS to contain "This is an auto-generated file for a definition", but it did not. It was actually %q`, js)
	}
}

func TestDefRepoNotFound(t *testing.T) {
	c, mock := apptest.New()

	mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	resp, err := c.Get(router.Rel.URLToDef(graph.DefKey{Repo: "my/repo", UnitType: "t", Unit: "u", Path: "p"}).String())
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}
}
