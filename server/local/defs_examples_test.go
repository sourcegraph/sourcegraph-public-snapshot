package local

import (
	"errors"
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestDefsService_ListExamples(t *testing.T) {
	var s defs
	ctx, mock := testContext()

	want := []*sourcegraph.Example{{Ref: graph.Ref{File: "f"}}}

	var calledListRefs, calledRepoTreeGet bool
	mock.servers.Defs.ListRefs_ = func(ctx context.Context, op *sourcegraph.DefsListRefsOp) (*sourcegraph.RefList, error) {
		calledListRefs = true
		return &sourcegraph.RefList{Refs: []*graph.Ref{
			&want[0].Ref,
		}}, nil
	}
	mock.servers.RepoTree.Get_ = func(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
		calledRepoTreeGet = true
		return &sourcegraph.TreeEntry{
			BasicTreeEntry: &sourcegraph.BasicTreeEntry{Type: sourcegraph.FileEntry},
			FileRange:      &sourcegraph.FileRange{},
		}, nil
	}

	exs, err := s.ListExamples(ctx, &sourcegraph.DefsListExamplesOp{
		Def: sourcegraph.DefSpec{CommitID: "c", Repo: "r", Path: "p"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exs.Examples, want) {
		t.Errorf("got %+v, want %+v", exs.Examples, want)
	}
	if !calledListRefs {
		t.Error("!calledListRefs")
	}
	if !calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}

// Simulate not being able to get the file from the VCS data store.
func TestDefsService_ListExamples_HandleTreeEntryError(t *testing.T) {
	var s defs
	ctx, mock := testContext()

	want := []*sourcegraph.Example{{Error: true, Ref: graph.Ref{File: "f"}}}

	var calledListRefs, calledRepoTreeGet bool
	mock.servers.Defs.ListRefs_ = func(ctx context.Context, op *sourcegraph.DefsListRefsOp) (*sourcegraph.RefList, error) {
		calledListRefs = true
		return &sourcegraph.RefList{Refs: []*graph.Ref{
			&want[0].Ref,
		}}, nil
	}
	mock.servers.RepoTree.Get_ = func(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
		calledRepoTreeGet = true
		return nil, errors.New("x")
	}

	exs, err := s.ListExamples(ctx, &sourcegraph.DefsListExamplesOp{
		Def: sourcegraph.DefSpec{CommitID: "c", Repo: "r", Path: "p"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exs.Examples, want) {
		t.Errorf("got %+v, want %+v", exs.Examples, want)
	}
	if !calledListRefs {
		t.Error("!calledListRefs")
	}
	if !calledRepoTreeGet {
		t.Error("!calledRepoTreeGet")
	}
}
