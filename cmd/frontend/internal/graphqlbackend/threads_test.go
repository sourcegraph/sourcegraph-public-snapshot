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
