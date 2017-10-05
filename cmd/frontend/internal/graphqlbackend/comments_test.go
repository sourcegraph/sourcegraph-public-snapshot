package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestComments_Create(t *testing.T) {
	t.Skip("broken due to test using the DB")

	ctx := context.Background()

	repo := sourcegraph.OrgRepo{
		ID:        1,
		RemoteURI: "github.com/foo/bar",
	}
	thread := sourcegraph.Thread{
		ID:             1,
		OrgRepoID:      1,
		File:           "foo.go",
		Revision:       "1234",
		StartLine:      1,
		EndLine:        2,
		StartCharacter: 3,
		EndCharacter:   4,
	}
	wantComment := sourcegraph.Comment{
		ThreadID:    1,
		Contents:    "Hello",
		AuthorName:  "Alice",
		AuthorEmail: "alice@acme.com",
	}

	store.Mocks.OrgRepos.MockGetByID_Return(t, &repo, nil)
	store.Mocks.Threads.MockGet_Return(t, &thread, nil)
	called, calledWith := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	_, err := r.AddCommentToThread(ctx, &struct {
		ThreadID int32
		Contents string
	}{
		ThreadID: thread.ID,
		Contents: wantComment.Contents,
	})

	if err != nil {
		t.Error(err)
	}
	if !*called || !reflect.DeepEqual(wantComment, *calledWith) {
		t.Errorf("want Comments.Create call to be %v not %v", wantComment, *calledWith)
	}
}

func TestComments_CreateAccessDenied(t *testing.T) {
	ctx := context.Background()

	store.Mocks.Threads.MockGet_Return(t, &sourcegraph.Thread{OrgRepoID: 1}, nil)
	store.Mocks.OrgRepos.MockGetByID_Return(t, nil, store.ErrRepoNotFound)
	called, calledWith := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	comment, err := r.AddCommentToThread(ctx, &struct {
		ThreadID int32
		Contents string
	}{})

	if *called {
		t.Errorf("should not have called Comments.Create (called with %v)", calledWith)
	}
	if comment != nil || err == nil {
		t.Error("did not return error for failed Comments.Create")
	}
}

func TestRepoNameFromURI(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "github.com/gorilla/mux", Out: "gorilla/mux"},
		{In: "git.acmeinternal.org/acme/acme", Out: "acme/acme"},
		{In: "company.internal/project", Out: "project"},
	}

	for _, test := range tests {
		out := repoNameFromURI(test.In)
		if out != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, out)
		}
	}
}
