package app_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"golang.org/x/net/context"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestDefPopover(t *testing.T) {
	c, mock := apptest.New()

	def := &sourcegraph.Def{Def: graph.Def{DefKey: graph.DefKey{Repo: "my/repo", UnitType: "GoPackage", Unit: "u", Path: "p"}}}

	calledReposGet := mockRepoGet(mock, "my/repo")
	mockCurrentRepoBuild(mock)
	mockEnabledRepoConfig(mock)
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	calledDefsGet := mock.Defs.MockGet_Return(t, def)

	if _, err := c.GetOK(router.Rel.URLToDefSubroute(router.DefPopover, def.DefKey).String()); err != nil {
		t.Fatal(err)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
	if !*calledDefsGet {
		t.Error("!calledDefsGet")
	}
}

func TestDefPopover_unbuiltDisplayEmpty(t *testing.T) {
	c, mock := apptest.New()

	calledReposGet := mockRepoGet(mock, "my/repo")
	calledReposGetCommit := mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mock.Defs.Get_ = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		return nil, errors.New("no def")
	}
	mockNoRepoBuild(mock)
	mockEnabledRepoConfig(mock)

	resp, err := c.Get(router.Rel.URLToDefSubroute(router.DefPopover, graph.DefKey{Repo: "my/repo", UnitType: "GoPackage", Unit: "u", Path: "p"}).String())
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusNotFound; resp.StatusCode != want {
		t.Errorf("got HTTP %d, want %d", resp.StatusCode, want)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "NoBuildError") {
		t.Errorf("didn't get correct response %q", b)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledReposGetCommit {
		t.Error("!calledReposGetCommit")
	}
}
