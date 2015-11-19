package local

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	vcstesting "sourcegraph.com/sourcegraph/go-vcs/vcs/testing"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestDefsService_ListAuthors_NoDB(t *testing.T) {
	var s defs
	ctx, mock := testContext()

	t1 := pbtypes.NewTimestamp(time.Unix(12345, 0))

	want := []*sourcegraph.DefAuthor{
		{
			UID: 1,
			DefAuthorship: sourcegraph.DefAuthorship{
				AuthorshipInfo: sourcegraph.AuthorshipInfo{
					LastCommitDate: t1,
					LastCommitID:   "c",
				},
				Bytes:           5,
				BytesProportion: 0.5,
			},
		},
		{
			Email: "a@a.com",
			DefAuthorship: sourcegraph.DefAuthorship{
				AuthorshipInfo: sourcegraph.AuthorshipInfo{
					LastCommitDate: t1,
					LastCommitID:   "c2",
				},
				Bytes:           5,
				BytesProportion: 0.5,
			},
		},
	}

	defSpec := sourcegraph.DefSpec{
		Repo:     "r",
		CommitID: "c",
		Unit:     "u",
		UnitType: "t",
		Path:     "p",
	}

	calledGet := mock.servers.Defs.MockGet_Return(t, &sourcegraph.Def{Def: graph.Def{
		DefKey:   defSpec.DefKey(),
		DefStart: 10,
		DefEnd:   20,
	}})
	var calledVCSRepoBlameFile bool
	mock.stores.RepoVCS.MockOpen(t, "r", vcstesting.MockRepository{
		BlameFile_: func(path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
			calledVCSRepoBlameFile = true
			return []*vcs.Hunk{
				{StartByte: 5, EndByte: 15, CommitID: "c", Author: vcs.Signature{Email: "u@u.com", Date: t1}},
				{StartByte: 15, EndByte: 25, CommitID: "c2", Author: vcs.Signature{Email: "a@a.com", Date: t1}},
			}, nil
		},
	})
	var calledDirectoryGetUserByEmail bool
	mock.stores.Directory.GetUserByEmail_ = func(ctx context.Context, email string) (*sourcegraph.UserSpec, error) {
		calledDirectoryGetUserByEmail = true
		if email == "u@u.com" {
			return &sourcegraph.UserSpec{UID: 1}, nil
		}
		return nil, nil
	}

	authors, err := s.ListAuthors(ctx, &sourcegraph.DefsListAuthorsOp{Def: defSpec})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(authors.DefAuthors, want) {
		t.Errorf("got %+v, want %+v", authors.DefAuthors, want)
	}
	if !*calledGet {
		t.Error("!calledGet")
	}
	if !calledVCSRepoBlameFile {
		t.Error("!calledVCSRepoBlameFile")
	}
	if !calledDirectoryGetUserByEmail {
		t.Error("!calledDirectoryGetUserByEmail")
	}
}
