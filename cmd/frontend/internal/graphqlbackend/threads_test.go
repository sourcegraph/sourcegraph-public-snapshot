package graphqlbackend

import (
	"context"
	"reflect"
	"testing"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func TestThreads_Create(t *testing.T) {
	ctx := context.Background()

	wantRepo := sourcegraph.LocalRepo{
		RemoteURI:   "test",
		AccessToken: "1234",
	}
	store.Mocks.LocalRepos.MockGet_Return(t, nil, store.ErrRepoNotFound)
	repoCreateCalled, repoCreateCalledWith := store.Mocks.LocalRepos.MockCreate_Return(t, &sourcegraph.LocalRepo{
		ID:          1,
		RemoteURI:   wantRepo.RemoteURI,
		AccessToken: wantRepo.AccessToken,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)
	threadCreateCalled, _ := store.Mocks.Threads.MockCreate_Return(t, &sourcegraph.Thread{
		ID:          1,
		LocalRepoID: wantRepo.ID,
		File:        "foo.go",
		Revision:    "abcd",
	}, nil)
	commentCreateCalled, _ := store.Mocks.Comments.MockCreate(t)

	r := &schemaResolver{}
	_, err := r.CreateThread(ctx, &struct {
		RemoteURI      string
		AccessToken    string
		File           string
		Revision       string
		StartLine      int32
		EndLine        int32
		StartCharacter int32
		EndCharacter   int32
		Contents       string
		AuthorName     string
		AuthorEmail    string
	}{
		RemoteURI:   wantRepo.RemoteURI,
		AccessToken: wantRepo.AccessToken,
		File:        "foo.go",
		Revision:    "abcd",
		Contents:    "Hello",
		AuthorName:  "Alice",
		AuthorEmail: "alice@acme.com",
	})

	if err != nil {
		t.Error(err)
	}
	if !*repoCreateCalled || !reflect.DeepEqual(wantRepo, *repoCreateCalledWith) {
		t.Errorf("want LocalRepos.Create call to be %v not %v", wantRepo, repoCreateCalledWith)
	}
	if !*threadCreateCalled {
		t.Error("expected Threads.Create to be called")
	}
	if !*commentCreateCalled {
		t.Error("expected Comments.Create to be called")
	}
}

func TestThreads_Get_RepoNotFound(t *testing.T) {
	ctx := context.Background()

	store.Mocks.LocalRepos.MockGet_Return(t, nil, store.ErrRepoNotFound)

	r := &rootResolver{}
	threads, err := r.Threads(ctx, &struct {
		RemoteURI   string
		AccessToken string
		File        string
	}{
		RemoteURI:   "test",
		AccessToken: "1234",
		File:        "foo.go",
	})

	if err != nil {
		t.Error(err)
	}

	if len(threads) != 0 {
		t.Errorf("expected threads to have length 0; got %#v", threads)
	}
}

func TestThreads_Update(t *testing.T) {
	wantRepo := sourcegraph.LocalRepo{
		RemoteURI:   "test",
		AccessToken: "1234",
	}

	store.Mocks.LocalRepos.MockGet_Return(t, &wantRepo, nil)
	called := store.Mocks.Threads.MockUpdate_Return(t, &sourcegraph.Thread{}, nil)

	r := &schemaResolver{}
	archived := true
	_, err := r.UpdateThread(context.Background(), &struct {
		RemoteURI   string
		AccessToken string
		ThreadID    int32
		Archived    *bool
	}{
		RemoteURI:   wantRepo.RemoteURI,
		AccessToken: wantRepo.AccessToken,
		ThreadID:    1,
		Archived:    &archived,
	})

	if err != nil {
		t.Error(err)
	}
	if !*called {
		t.Error("expected Threads.Update to be called")
	}
}

func TestTitleFromContents(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "Hello", Out: "Hello"},
		{In: "Hello.", Out: "Hello."},
		{In: "Hello?", Out: "Hello?"},
		{In: "Hello!", Out: "Hello!"},
		{In: "Hello there!", Out: "Hello there!"},
		{In: "Check this out. Weird code huh?", Out: "Check this out."},
		{In: "Hello world\n", Out: "Hello world"},
		{In: "Hello world\nAnd all who inhabit it.", Out: "Hello world"},
		{In: "Hello title\n\nSome contents?", Out: "Hello title"},
		{In: "I have a question about this.\nWhat's going on here", Out: "I have a question about this."},
		{In: "Hello title\n\nSome contents?", Out: "Hello title"},
		{In: "What does foo.bar do?", Out: "What does foo.bar do?"},
		{In: "It should be 1 != 2\nFYI 1 != 1 is wrong.", Out: "It should be 1 != 2"},
	}

	for _, test := range tests {
		out := titleFromContents(test.In)
		if out != test.Out {
			t.Errorf("\n   input: \"%s\"\nexpected: \"%s\"\n     got: \"%s\"", test.In, test.Out, out)
		}
	}
}
