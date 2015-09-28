package app_test

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/apptest"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestRepoMain_redirectToNewURI(t *testing.T) {
	c, mock := apptest.New()

	const (
		oldURI = "old/repo"
		newURI = "new/repo"
	)

	var calledGet bool
	mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet = true
		if repo.URI != oldURI {
			t.Errorf("got %q, want %q", repo.URI, oldURI)
		}
		return &sourcegraph.Repo{URI: newURI, DefaultBranch: "master"}, nil
	}
	mockEnabledRepoConfig(mock)

	err := checkRedirection(c,
		router.Rel.URLToRepo("old/repo").String(),
		router.Rel.URLToRepo("new/repo").String(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !calledGet {
		t.Error("!calledGet")
	}
}

func TestRepoSubroute_redirectToNewURI(t *testing.T) {
	c, mock := apptest.New()

	const (
		oldURI = "old/repo"
		newURI = "new/repo"
	)

	var calledGet bool
	mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet = true
		if repo.URI != oldURI {
			t.Errorf("got %q, want %q", repo.URI, oldURI)
		}
		return &sourcegraph.Repo{URI: newURI, DefaultBranch: "master"}, nil
	}
	mockEnabledRepoConfig(mock)

	err := checkRedirection(c,
		router.Rel.URLToRepoSubroute(router.RepoBranches, "old/repo").String(),
		router.Rel.URLToRepoSubroute(router.RepoBranches, "new/repo").String(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !calledGet {
		t.Error("!calledGet")
	}
}

func TestDef_redirectToNewURI(t *testing.T) {
	c, mock := apptest.New()

	const (
		oldURI = "old/repo"
		newURI = "new/repo"
	)

	var calledGet bool
	mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		calledGet = true
		if repo.URI != oldURI {
			t.Errorf("got %q, want %q", repo.URI, oldURI)
		}
		return &sourcegraph.Repo{URI: newURI, DefaultBranch: "master"}, nil
	}
	mockEnabledRepoConfig(mock)

	oldDefKey := graph.DefKey{Repo: oldURI, UnitType: "t", Unit: "u", Path: "p"}
	newDefKey := graph.DefKey{Repo: newURI, UnitType: "t", Unit: "u", Path: "p"}

	err := checkRedirection(c,
		router.Rel.URLToDef(oldDefKey).String(),
		router.Rel.URLToDef(newDefKey).String(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if !calledGet {
		t.Error("!calledGet")
	}
}

func checkRedirection(c *httptestutil.Client, fromURL, toURL string) error {
	resp, err := c.GetNoFollowRedirects(fromURL)
	if err != nil {
		return err
	}

	if resp.StatusCode < 300 || resp.StatusCode > 399 {
		return fmt.Errorf("got HTTP %d, want 300-399 (redirect)", resp.StatusCode)
	}
	if dest := resp.Header.Get("location"); toURL != dest {
		return fmt.Errorf("got redirect to %q, want %q", toURL, dest)
	}
	return nil
}
