package background

import (
	"context"
	"testing"

	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore/mocks"
)

func TestCommitUpdater(t *testing.T) {
	graph := map[string][]string{
		"a": nil,
		"b": {"a"},
	}

	mockStore := storemocks.NewMockStore()
	mockStore.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockStore.LockFunc.SetDefaultReturn(false, nil, nil)

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.CommitGraphFunc.SetDefaultReturn(graph, nil)
	mockGitserverClient.HeadFunc.SetDefaultReturn("b", nil)

	updater := &CommitUpdater{
		store:           mockStore,
		gitserverClient: mockGitserverClient,
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error updating commit graph: %s", err)
	}

	if len(mockStore.LockFunc.History()) != 1 {
		t.Fatalf("unexpected lock call count. want=%d have=%d", 1, len(mockStore.LockFunc.History()))
	} else {
		call := mockStore.LockFunc.History()[0]
		if call.Arg1 != 42 {
			t.Errorf("unexpected repository id argument. want=%d have=%d", 42, call.Arg1)
		}
		if call.Arg2 {
			t.Errorf("unexpected blocking argument. want=%v have=%v", false, call.Arg2)
		}
	}

	if len(mockStore.CalculateVisibleUploadsFunc.History()) != 0 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 0, len(mockStore.CalculateVisibleUploadsFunc.History()))
	}
}
