package commitgraph

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestUpdater(t *testing.T) {
	graph := gitserver.ParseCommitGraph([]string{
		"a",
		"b a",
	})

	commitTime := time.Unix(1587396557, 0).UTC()
	mockDBStore := NewMockDBStore()
	mockDBStore.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockDBStore.GetOldestCommitDateFunc.SetDefaultReturn(commitTime, true, nil)

	mockLocker := NewMockLocker()
	mockLocker.LockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.CommitGraphFunc.SetDefaultReturn(graph, nil)
	mockGitserverClient.HeadFunc.SetDefaultReturn("b", nil)

	updater := &Updater{
		dbStore:         mockDBStore,
		locker:          mockLocker,
		gitserverClient: mockGitserverClient,
		operations:      newOperations(mockDBStore, &observation.TestContext),
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error updating commit graph: %s", err)
	}

	if len(mockLocker.LockFunc.History()) != 1 {
		t.Fatalf("unexpected lock call count. want=%d have=%d", 1, len(mockLocker.LockFunc.History()))
	} else {
		call := mockLocker.LockFunc.History()[0]
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

func TestUpdaterNoUploads(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockDBStore.GetOldestCommitDateFunc.SetDefaultReturn(time.Time{}, false, nil)
	mockLocker := NewMockLocker()
	mockLocker.LockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)
	mockGitserverClient := NewMockGitserverClient()

	updater := &Updater{
		dbStore:         mockDBStore,
		locker:          mockLocker,
		gitserverClient: mockGitserverClient,
		operations:      newOperations(mockDBStore, &observation.TestContext),
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

func TestUpdaterLocked(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockLocker := NewMockLocker()
	mockLocker.LockFunc.SetDefaultReturn(false, nil, nil)
	mockGitserverClient := NewMockGitserverClient()

	updater := &Updater{
		dbStore:         mockDBStore,
		locker:          mockLocker,
		gitserverClient: mockGitserverClient,
		operations:      newOperations(mockDBStore, &observation.TestContext),
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error updating commit graph: %s", err)
	}

	if len(mockDBStore.CalculateVisibleUploadsFunc.History()) != 0 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 0, len(mockDBStore.CalculateVisibleUploadsFunc.History()))
	}
}
