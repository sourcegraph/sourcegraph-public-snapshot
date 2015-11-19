package app_test

import (
	"strings"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
)

func TestRepoReadme(t *testing.T) {
	c, mock := apptest.New()

	const wantReadmeText = "hello from the readme"

	mockBasicRepoMainPage(mock)
	mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.Repos.GetReadme_ = func(ctx context.Context, repo *sourcegraph.RepoRevSpec) (*sourcegraph.Readme, error) {
		return &sourcegraph.Readme{Path: "README", HTML: wantReadmeText}, nil
	}

	resp, err := c.GetOK(router.Rel.URLToRepo("my/repo").String())
	if err != nil {
		t.Fatal(err)
	}
	dom, err := parseHTML(resp)
	if err != nil {
		t.Fatal(err)
	}

	if readmeText := dom.Find(".readme").Text(); !strings.Contains(readmeText, wantReadmeText) {
		t.Errorf("got readme text %q, want %q", readmeText, wantReadmeText)
	}

	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}

func TestRepoReadme_NoReadme(t *testing.T) {
	c, mock := apptest.New()

	mockBasicRepoMainPage(mock)
	mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.Repos.GetReadme_ = func(ctx context.Context, repo *sourcegraph.RepoRevSpec) (*sourcegraph.Readme, error) {
		return nil, grpc.Errorf(codes.NotFound, "")
	}

	resp, err := c.GetOK(router.Rel.URLToRepo("my/repo").String())
	if err != nil {
		t.Fatal(err)
	}
	dom, err := parseHTML(resp)
	if err != nil {
		t.Fatal(err)
	}

	if readmeElem := dom.Find(".readme"); readmeElem.Size() != 0 {
		t.Errorf("got a .readme element, but want none")
	}

	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}
