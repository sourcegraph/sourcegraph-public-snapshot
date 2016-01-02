package app_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestDefPopover(t *testing.T) {
	c, mock := apptest.New()

	def := &sourcegraph.Def{Def: graph.Def{DefKey: graph.DefKey{Repo: "my/repo", CommitID: "c", UnitType: "GoPackage", Unit: "u", Path: "p"}}}

	calledReposGet := mockRepoGet(mock, "my/repo")
	mockSpecificVersionSrclibData(mock, "c")
	mockEmptyRepoConfig(mock)
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
