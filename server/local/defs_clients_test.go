package local

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestDefsService_ListClients(t *testing.T) {
	var s defs
	ctx, mock := testContext()

	t1 := pbtypes.NewTimestamp(time.Unix(12345, 0).In(time.UTC))

	want := []*sourcegraph.DefClient{
		{
			UID: 1,
			AuthorshipInfo: sourcegraph.AuthorshipInfo{
				LastCommitDate: t1,
				LastCommitID:   "c",
			},
		},
		{
			Email: "a@a.com",
			AuthorshipInfo: sourcegraph.AuthorshipInfo{
				LastCommitDate: t1,
				LastCommitID:   "c2",
			},
		},
	}

	defSpec := sourcegraph.DefSpec{Repo: "r", CommitID: "c", UnitType: "t", Unit: "u", Path: "p"}

	mock.servers.Defs.ListRefs_ = func(ctx context.Context, op *sourcegraph.DefsListRefsOp) (*sourcegraph.RefList, error) {
		return &sourcegraph.RefList{Refs: []*sourcegraph.Ref{
			{
				Ref: graph.Ref{
					DefRepo: "r", DefUnitType: "t", DefUnit: "u", DefPath: "p", Repo: "r2",
					CommitID: "c", File: "f2", Start: 10, End: 20,
				},
				Authorship: &sourcegraph.AuthorshipInfo{AuthorEmail: "a@a.com", LastCommitID: "c2", LastCommitDate: t1},
			},
			{
				Ref: graph.Ref{
					DefRepo: "r", DefUnitType: "t", DefUnit: "u", DefPath: "p", Repo: "r3",
					CommitID: "c", File: "f3", Start: 20, End: 30,
				},
				Authorship: &sourcegraph.AuthorshipInfo{AuthorEmail: "u@u.com", LastCommitID: "c", LastCommitDate: t1},
			},
		}}, nil
	}
	var calledDirectoryGetUserByEmail bool
	mock.stores.Directory.GetUserByEmail_ = func(ctx context.Context, email string) (*sourcegraph.UserSpec, error) {
		calledDirectoryGetUserByEmail = true
		if email == "u@u.com" {
			return &sourcegraph.UserSpec{UID: 1}, nil
		}
		return nil, nil
	}

	clients, err := s.ListClients(ctx, &sourcegraph.DefsListClientsOp{Def: defSpec})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(clients.DefClients, want) {
		t.Errorf("got %+v, want %+v", clients.DefClients, want)
	}
	if !calledDirectoryGetUserByEmail {
		t.Error("!calledDirectoryGetUserByEmail")
	}
}
