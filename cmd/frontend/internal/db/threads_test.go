package db

import (
	"fmt"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

func TestThreads_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	user, err := Users.Create(ctx, NewUser{
		Email:     "a@a.com",
		Username:  "u",
		Password:  "p",
		EmailCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	thread, err := Threads.Create(ctx, &types.Thread{AuthorUserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}

	if count, err := Threads.Count(ctx, ThreadsListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	if err := Threads.Delete(ctx, thread.ID); err != nil {
		t.Fatal(err)
	}

	if count, err := Threads.Count(ctx, ThreadsListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
}

func TestThreads_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	user, err := Users.Create(ctx, NewUser{
		Email:     "a@a.com",
		Username:  "u",
		Password:  "p",
		EmailCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	thread, err := Threads.Create(ctx, &types.Thread{AuthorUserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}

	// Delete thread.
	if err := Threads.Delete(ctx, thread.ID); err != nil {
		t.Fatal(err)
	}

	// Thread no longer exists.
	_, err = Threads.Get(ctx, thread.ID)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("thread %d not found", thread.ID)) {
		t.Errorf("got error %v, want thread not found", err)
	}
	orgs, err := Threads.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(orgs) > 0 {
		t.Errorf("got %d orgs, want 0", len(orgs))
	}

	// Can't delete already-deleted thread.
	err = Threads.Delete(ctx, thread.ID)
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("thread %d not found", thread.ID)) {
		t.Errorf("got error %v, want thread not found", err)
	}
}
