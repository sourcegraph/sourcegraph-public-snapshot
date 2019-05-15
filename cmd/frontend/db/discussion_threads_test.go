package db

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

// TODO(slimsag:discussions): future: test that DiscussionThreadsListOptions.AuthorUserID works

func TestDiscussionThreads_CreateAddTargetGet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create the thread.
	thread, err := DiscussionThreads.Create(ctx, &types.DiscussionThread{
		AuthorUserID: user.ID,
		Title:        "Hello world!",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get the thread we just created.
	gotThread, err := DiscussionThreads.Get(ctx, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotThread.ID != thread.ID {
		t.Logf("got thread ID:  %v", gotThread.ID)
		t.Fatalf("want thread ID: %v", thread.ID)
	}
	if gotThread.AuthorUserID != thread.AuthorUserID {
		t.Logf("got thread AuthorUserID:  %v", gotThread.AuthorUserID)
		t.Fatalf("want thread AuthorUserID: %v", thread.AuthorUserID)
	}
}

func TestDiscussionThreads_Update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create the thread.
	thread, err := DiscussionThreads.Create(ctx, &types.DiscussionThread{
		AuthorUserID: user.ID,
		Title:        "Hello world!",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Update the thread.
	const wantTitle = "x"
	gotThread, err := DiscussionThreads.Update(ctx, thread.ID, &DiscussionThreadsUpdateOptions{
		Title:   strPtr(wantTitle),
		Archive: boolPtr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotThread.Title != wantTitle {
		t.Errorf("got title %q, want %q", gotThread.Title, wantTitle)
	}
	if gotThread.ArchivedAt == nil {
		t.Fatal("expected thread to be archived")
	}
}

func TestDiscussionThreads_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create the thread.
	thread, err := DiscussionThreads.Create(ctx, &types.DiscussionThread{
		AuthorUserID: user.ID,
		Title:        "Hello world!",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Count threads.
	count, err := DiscussionThreads.Count(ctx, &DiscussionThreadsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d, want 1", count)
	}

	// Delete the thread.
	if _, err := DiscussionThreads.Update(ctx, thread.ID, &DiscussionThreadsUpdateOptions{Delete: true}); err != nil {
		t.Fatal(err)
	}

	// Count threads.
	count, err = DiscussionThreads.Count(ctx, &DiscussionThreadsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d, want 0", count)
	}
}

func TestDiscussionThreads_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create the thread.
	thread, err := DiscussionThreads.Create(ctx, &types.DiscussionThread{
		AuthorUserID: user.ID,
		Title:        "Hello world!",
	})
	if err != nil {
		t.Fatal(err)
	}

	// List threads.
	threads, err := DiscussionThreads.List(ctx, &DiscussionThreadsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 1 {
		t.Fatalf("got %d threads, want 1", len(threads))
	}

	// Delete the thread.
	if _, err := DiscussionThreads.Update(ctx, thread.ID, &DiscussionThreadsUpdateOptions{Delete: true}); err != nil {
		t.Fatal(err)
	}

	// List threads.
	threads, err = DiscussionThreads.List(ctx, &DiscussionThreadsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 0 {
		t.Fatalf("got %d threads, want 0", len(threads))
	}
}

func TestDiscussionThreads_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create the thread.
	thread, err := DiscussionThreads.Create(ctx, &types.DiscussionThread{
		AuthorUserID: user.ID,
		Title:        "Hello world!",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Delete thread.
	if _, err := DiscussionThreads.Update(ctx, thread.ID, &DiscussionThreadsUpdateOptions{Delete: true}); err != nil {
		t.Fatal(err)
	}

	// Thread no longer exists.
	_, err = DiscussionThreads.Get(ctx, thread.ID)
	if _, ok := err.(*ErrThreadNotFound); !ok {
		t.Errorf("got error %v, want thread not found", err)
	}

	// List threads.
	threads, err := DiscussionThreads.List(ctx, &DiscussionThreadsListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 0 {
		t.Fatalf("got %d threads, want 0", len(threads))
	}

	// Deleting an already-deleted thread should be no-op.
	updatedThread, err := DiscussionThreads.Update(ctx, thread.ID, &DiscussionThreadsUpdateOptions{Delete: true})
	if updatedThread != nil || err != nil {
		t.Errorf("got updatedThread=%v err=%v, want nil thread nil error", updatedThread, err)
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}
