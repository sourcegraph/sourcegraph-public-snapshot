package background

import (
	"context"
	"testing"
	"time"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
)

func TestCommitUpdater(t *testing.T) {
	graph := map[string][]string{
		"a": nil,
		"b": {"a"},
	}

	mockDBStore := NewMockDBStore()
	mockDBStore.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockDBStore.LockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)
	mockDBStore.OldestDumpForRepositoryFunc.SetDefaultReturn(store.Dump{Commit: "deadbeef"}, true, nil)

	commitTime := time.Unix(1587396557, 0).UTC()
	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.CommitGraphFunc.SetDefaultReturn(graph, nil)
	mockGitserverClient.CommitDateFunc.SetDefaultReturn(commitTime, nil)
	mockGitserverClient.HeadFunc.SetDefaultReturn("b", nil)

	updater := &CommitUpdater{
		dbStore:         mockDBStore,
		gitserverClient: mockGitserverClient,
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error updating commit graph: %s", err)
	}

	if len(mockDBStore.LockFunc.History()) != 1 {
		t.Fatalf("unexpected lock call count. want=%d have=%d", 1, len(mockDBStore.LockFunc.History()))
	} else {
		call := mockDBStore.LockFunc.History()[0]
		if call.Arg1 != 42 {
			t.Errorf("unexpected repository id argument. want=%d have=%d", 42, call.Arg1)
		}
		if call.Arg2 {
			t.Errorf("unexpected blocking argument. want=%v have=%v", false, call.Arg2)
		}
	}

	// Should fetch commit graph
	if len(mockGitserverClient.CommitGraphFunc.History()) != 1 {
		t.Fatalf("unexpected commit graph call count. want=%d have=%d", 1, len(mockGitserverClient.CommitGraphFunc.History()))
	}
	// Should calculate visible uploads with fetched graph
	if len(mockDBStore.CalculateVisibleUploadsFunc.History()) != 1 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 1, len(mockDBStore.CalculateVisibleUploadsFunc.History()))
	}
}

func TestCommitUpdaterNoOldDump(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockDBStore.LockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)
	mockDBStore.OldestDumpForRepositoryFunc.SetDefaultReturn(store.Dump{}, false, nil)
	mockGitserverClient := NewMockGitserverClient()

	updater := &CommitUpdater{
		dbStore:         mockDBStore,
		gitserverClient: mockGitserverClient,
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error updating commit graph: %s", err)
	}

	// Should not not fetch commit graph
	if len(mockGitserverClient.CommitGraphFunc.History()) != 0 {
		t.Fatalf("unexpected commit graph call count. want=%d have=%d", 0, len(mockGitserverClient.CommitGraphFunc.History()))
	}
	// Should calculate visible uploads with empty graph
	if len(mockDBStore.CalculateVisibleUploadsFunc.History()) != 1 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 1, len(mockDBStore.CalculateVisibleUploadsFunc.History()))
	}
}

func TestCommitUpdaterLocked(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockDBStore.LockFunc.SetDefaultReturn(false, nil, nil)
	mockGitserverClient := NewMockGitserverClient()

	updater := &CommitUpdater{
		dbStore:         mockDBStore,
		gitserverClient: mockGitserverClient,
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error updating commit graph: %s", err)
	}

	if len(mockDBStore.CalculateVisibleUploadsFunc.History()) != 0 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 0, len(mockDBStore.CalculateVisibleUploadsFunc.History()))
	}
}
