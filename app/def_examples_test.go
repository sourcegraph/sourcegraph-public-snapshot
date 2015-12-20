package app_test

import (
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestDefExamples(t *testing.T) {
	c, mock := apptest.New()

	def := &sourcegraph.Def{
		Def: graph.Def{
			DefKey: graph.DefKey{Repo: "my/repo", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"},
			Kind:   "func",
			File:   "a.txt",
		},
	}

	calledReposGet := mockRepoGet(mock, "my/repo")
	mock.Repos.MockGetCommit_ByID_NoCheck(t, "c")
	mockSpecificVersionSrclibData(mock, "c")
	mockEmptyRepoConfig(mock)
	calledDefsGet := mock.Defs.MockGet_Return(t, def)

	var calledDefsListExamples bool
	mock.Defs.ListExamples_ = func(ctx context.Context, op *sourcegraph.DefsListExamplesOp) (*sourcegraph.ExampleList, error) {
		calledDefsListExamples = true
		return &sourcegraph.ExampleList{
			Examples: []*sourcegraph.Example{
				{
					Ref:     graph.Ref{DefRepo: "my/repo", DefUnitType: "t", DefUnit: "u", DefPath: "p", Repo: "r", File: "foo.go", CommitID: "c", Start: 3, End: 6},
					SrcHTML: `<a>baz</a>`,
				},
				{
					Ref:     graph.Ref{DefRepo: "my/repo", DefUnitType: "t", DefUnit: "u", DefPath: "p", Repo: "r", File: "bar.go", CommitID: "c", Start: 20, End: 23},
					SrcHTML: `<a>baz</a>`,
				},
			},
		}, nil
	}

	if _, err := c.GetOK(router.Rel.URLToDefAtRevSubroute(router.DefExamples, def.DefKey, def.CommitID).String()); err != nil {
		t.Fatal(err)
	}

	if !*calledReposGet {
		t.Error("!calledReposGet")
	}
	if !*calledDefsGet {
		t.Error("!calledDefsGet")
	}
	if !calledDefsListExamples {
		t.Error("!calledDefsListExamples")
	}
}
