package background

import (
	"context"
	"testing"
)

func TestCommitUpdater(t *testing.T) {
	graph := map[string][]string{
		"a": nil,
		"b": {"a"},
	}

	mockDBStore := NewMockDBStore()
	mockDBStore.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockDBStore.LockFunc.SetDefaultReturn(false, nil, nil)

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.CommitGraphFunc.SetDefaultReturn(graph, nil)
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

	if len(mockDBStore.CalculateVisibleUploadsFunc.History()) != 0 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 0, len(mockDBStore.CalculateVisibleUploadsFunc.History()))
	}
}
