package httpapi

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestDeltaFiles_ok(t *testing.T) {
	c, mock := newTest()

	want := &sourcegraph.DeltaFiles{Stats: diff.Stat{Added: 1}}

	calledReposResolve := mock.Repos.MockResolve_Local(t, "r", 1)
	var calledReposResolveRevBase, calledReposResolveRevHead bool
	mock.Repos.ResolveRev_ = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
		if want := int32(1); op.Repo != want {
			t.Errorf("got repo %d, want %d", op.Repo, want)
		}
		switch op.Rev {
		case "vbase":
			calledReposResolveRevBase = true
			return &sourcegraph.ResolvedRev{CommitID: "cbase"}, nil
		case "vhead":
			calledReposResolveRevHead = true
			return &sourcegraph.ResolvedRev{CommitID: "chead"}, nil
		default:
			t.Fatalf("got rev %q, want \"vbase\" or \"vhead\"", op.Rev)
			panic("unreachable")
		}
	}
	var calledDeltasListFiles bool
	mock.Deltas.ListFiles_ = func(ctx context.Context, op *sourcegraph.DeltasListFilesOp) (*sourcegraph.DeltaFiles, error) {
		calledDeltasListFiles = true
		wantOp := sourcegraph.DeltasListFilesOp{
			Ds: sourcegraph.DeltaSpec{
				Base: sourcegraph.RepoRevSpec{Repo: 1, CommitID: "cbase"},
				Head: sourcegraph.RepoRevSpec{Repo: 1, CommitID: "chead"},
			},
			Opt: &sourcegraph.DeltaListFilesOptions{Filter: "f"},
		}
		if !reflect.DeepEqual(*op, wantOp) {
			t.Fatalf("got op %#v, want %#v", *op, wantOp)
		}
		return want, nil
	}

	var files *sourcegraph.DeltaFiles
	if err := c.GetJSON("/repos/r@vhead/-/delta/vbase/-/files?Filter=f", &files); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("got %+v, want %+v", files, want)
	}
	if !*calledReposResolve {
		t.Error("!calledReposResolve")
	}
	if !calledReposResolveRevBase {
		t.Error("!calledReposResolveRevBase")
	}
	if !calledReposResolveRevHead {
		t.Error("!calledReposResolveRevHead")
	}
	if !calledDeltasListFiles {
		t.Error("!calledDeltasListFiles")
	}
}
