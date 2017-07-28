package graphqlbackend

import (
	"context"
	"reflect"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestComments_Create(t *testing.T) {
	ctx := context.Background()

	repo := sourcegraph.LocalRepo{
		ID:          1,
		RemoteURI:   "github.com/foo/bar",
		AccessToken: "1234",
	}
	thread := sourcegraph.Thread{
		ID:             1,
		LocalRepoID:    1,
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

	store.Mocks.LocalRepos.MockGet_Return(t, &repo, nil)
	store.Mocks.Threads.MockGet_Return(t, &thread, nil)
	called, calledWith := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	_, err := r.AddCommentToThread(ctx, &struct {
		RemoteURI   string
		AccessToken string
		ThreadID    int32
		Contents    string
		AuthorName  string
		AuthorEmail string
	}{
		RemoteURI:   repo.RemoteURI,
		AccessToken: repo.AccessToken,
		ThreadID:    thread.ID,
		Contents:    wantComment.Contents,
		AuthorName:  wantComment.AuthorName,
		AuthorEmail: wantComment.AuthorEmail,
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

	store.Mocks.LocalRepos.MockGet_Return(t, nil, store.ErrRepoNotFound)
	called, calledWith := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	comment, err := r.AddCommentToThread(ctx, &struct {
		RemoteURI   string
		AccessToken string
		ThreadID    int32
		Contents    string
		AuthorName  string
		AuthorEmail string
	}{})

	if *called {
		t.Errorf("should not have called Comments.Create (called with %v)", calledWith)
	}
	if comment != nil || err == nil {
		t.Error("did not return error for failed Comments.Create")
	}
}
