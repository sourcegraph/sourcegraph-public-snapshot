package background

import (
	"testing"
	"time"

	commitsmocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/commits/mocks"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
)

func TestCommitUpdater(t *testing.T) {
	store := storemocks.NewMockStore()
	updater := commitsmocks.NewMockUpdater()
	store.DirtyRepositoriesFunc.SetDefaultReturn(map[int]int{50: 3, 51: 2, 52: 6}, nil)

	commitUpdater := NewCommitUpdater(store, updater, time.Second)
	go func() { commitUpdater.Start() }()
	commitUpdater.Stop()

	if callCount := len(updater.TryUpdateFunc.History()); callCount < 3 {
		t.Fatalf("unexpected update call count. want>=%d have=%d", 3, callCount)
	}

	testCases := []struct {
		repositoryID int
		dirtyFlag    int
	}{
		{50, 3},
		{51, 2},
		{52, 6},
	}
	for _, testCase := range testCases {
		found := false
		for _, call := range updater.TryUpdateFunc.History() {
			if call.Arg1 == testCase.repositoryID && call.Arg2 == testCase.dirtyFlag {
				found = true
			}
		}

		if !found {
			t.Errorf("unexpected call with args (%d, %d)", testCase.repositoryID, testCase.dirtyFlag)
		}
	}
}
