package backend

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstesting "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
)

func TestDefsService_ListAuthors(t *testing.T) {
	orig := feature.Features.Authors
	feature.Features.Authors = true
	defer func() {
		feature.Features.Authors = orig
	}()

	var s defs
	ctx, mock := testContext()

	t1 := pbtypes.NewTimestamp(time.Unix(12345, 0))

	want := []*sourcegraph.DefAuthor{
		{
			Email: "u",
			DefAuthorship: sourcegraph.DefAuthorship{
				AuthorshipInfo: sourcegraph.AuthorshipInfo{
					LastCommitDate: t1,
					LastCommitID:   "c",
				},
				Bytes:           6,
				BytesProportion: 0.6,
			},
		},
		{
			Email: "a",
			DefAuthorship: sourcegraph.DefAuthorship{
				AuthorshipInfo: sourcegraph.AuthorshipInfo{
					LastCommitDate: t1,
					LastCommitID:   "c2",
				},
				Bytes:           4,
				BytesProportion: 0.4,
			},
		},
	}

	defSpec := sourcegraph.DefSpec{
		Repo:     "r",
		CommitID: strings.Repeat("c", 40),
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
				{StartByte: 5, EndByte: 16, CommitID: "c", Author: vcs.Signature{Email: "u@u.com", Date: t1}},
				{StartByte: 16, EndByte: 25, CommitID: "c2", Author: vcs.Signature{Email: "a@a.com", Date: t1}},
			}, nil
		},
	})

	authors, err := s.ListAuthors(ctx, &sourcegraph.DefsListAuthorsOp{Def: defSpec})
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range authors.DefAuthors {
		a.AvatarURL = ""
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
}
