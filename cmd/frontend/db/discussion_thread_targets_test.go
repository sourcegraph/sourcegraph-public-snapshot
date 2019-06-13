package db

import (
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestDiscussionThreads_Targets(t *testing.T) {
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

	// Create a repository to comply with the postgres repo constraint.
	if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo, err := Repos.GetByName(ctx, "myrepo")
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

	// Add target 1.
	target1, err := DiscussionThreads.AddTarget(ctx, &types.DiscussionThreadTargetRepo{
		ThreadID: thread.ID,
		RepoID:   repo.ID,
		Path:     strPtr("foo/bar/mux.go"),
		Branch:   strPtr("master"),
		Revision: strPtr("0c1a96370c1a96370c1a96370c1a96370c1a9637"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Add target 2.
	target2, err := DiscussionThreads.AddTarget(ctx, &types.DiscussionThreadTargetRepo{
		ThreadID: thread.ID,
		RepoID:   repo.ID,
		Path:     strPtr("foo/qux.go"),
		Branch:   strPtr("master"),
		Revision: strPtr("0c1a96370c1a96370c1a96370c1a96370c1a9637"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get target 1.
	{
		gotTarget1, err := DiscussionThreads.GetTarget(ctx, target1.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotTarget1, target1) {
			t.Fatalf("got target1 %v, want %v", gotTarget1, target1)
		}
	}
	// Get target 2.
	{
		gotTarget2, err := DiscussionThreads.GetTarget(ctx, target2.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotTarget2, target2) {
			t.Fatalf("got target1 %v, want %v", gotTarget2, target2)
		}
	}

	// List targets.
	targets, err := DiscussionThreads.ListTargets(ctx, DiscussionThreadsListTargetsOptions{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].ID < targets[j].ID })
	if want := []*types.DiscussionThreadTargetRepo{target1, target2}; !reflect.DeepEqual(targets, want) {
		t.Errorf("got targets %v, want %v", targets, want)
	}

	// Remove a target.
	if err := DiscussionThreads.RemoveTarget(ctx, target1.ID); err != nil {
		t.Fatal(err)
	}
	targets, err = DiscussionThreads.ListTargets(ctx, DiscussionThreadsListTargetsOptions{ThreadID: thread.ID})
	if err != nil {
		t.Fatal(err)
	}
	if want := []*types.DiscussionThreadTargetRepo{target2}; !reflect.DeepEqual(targets, want) {
		t.Errorf("got targets %v, want %v", targets, want)
	}

	// Ensure getting target 1 returns nil (because it was removed).
	{
		gotTarget1, err := DiscussionThreads.GetTarget(ctx, target1.ID)
		if err != nil {
			t.Fatal(err)
		}
		if gotTarget1 != nil {
			t.Fatalf("got target1 %v, want nil", gotTarget1)
		}
	}
	// Get target 2.
	{
		gotTarget2, err := DiscussionThreads.GetTarget(ctx, target2.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(gotTarget2, target2) {
			t.Fatalf("got target1 %v, want %v", gotTarget2, target2)
		}
	}

}
