package commitgraph

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestUpdater(t *testing.T) {
	graph := gitdomain.ParseCommitGraph([]string{
		"a",
		"b a",
	})

	commitTime := time.Unix(1587396557, 0).UTC()
	mockDBStore := NewMockDBStore()
	mockUploadSvc := NewMockUploadService()
	mockUploadSvc.GetDirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockUploadSvc.GetOldestCommitDateFunc.SetDefaultReturn(commitTime, true, nil)

	mockLocker := NewMockLocker()
	mockLocker.LockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.CommitGraphFunc.SetDefaultReturn(graph, nil)
	mockGitserverClient.RefDescriptionsFunc.SetDefaultReturn(map[string][]gitdomain.RefDescription{
		"b": {{IsDefaultBranch: true}},
	}, nil)

	updater := &updater{
		dbStore:         mockDBStore,
		uploadSvc:       mockUploadSvc,
		locker:          mockLocker,
		gitserverClient: mockGitserverClient,
		operations:      NewOperations(mockDBStore, mockUploadSvc, &observation.TestContext),
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
	if len(mockUploadSvc.UpdateUploadsVisibleToCommitsFunc.History()) != 1 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 1, len(mockUploadSvc.UpdateUploadsVisibleToCommitsFunc.History()))
	}
}

func TestUpdaterNoUploads(t *testing.T) {
	mockDBStore := NewMockDBStore()

	mockUploadSvc := NewMockUploadService()
	mockUploadSvc.GetDirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)
	mockUploadSvc.GetOldestCommitDateFunc.SetDefaultReturn(time.Time{}, false, nil)

	mockLocker := NewMockLocker()
	mockLocker.LockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.RefDescriptionsFunc.SetDefaultReturn(map[string][]gitdomain.RefDescription{
		"b": {{IsDefaultBranch: true}},
	}, nil)

	updater := &updater{
		dbStore:         mockDBStore,
		uploadSvc:       mockUploadSvc,
		locker:          mockLocker,
		gitserverClient: mockGitserverClient,
		operations:      NewOperations(mockDBStore, mockUploadSvc, &observation.TestContext),
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error updating commit graph: %s", err)
	}

	// Should not not fetch commit graph
	if len(mockGitserverClient.CommitGraphFunc.History()) != 0 {
		t.Fatalf("unexpected commit graph call count. want=%d have=%d", 0, len(mockGitserverClient.CommitGraphFunc.History()))
	}
	// Should calculate visible uploads with empty graph
	if len(mockUploadSvc.UpdateUploadsVisibleToCommitsFunc.History()) != 1 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 1, len(mockUploadSvc.UpdateUploadsVisibleToCommitsFunc.History()))
	}
}

func TestUpdaterLocked(t *testing.T) {
	mockUploadSvc := NewMockUploadService()
	mockUploadSvc.GetDirtyRepositoriesFunc.SetDefaultReturn(map[int]int{42: 15}, nil)

	mockDBStore := NewMockDBStore()
	mockLocker := NewMockLocker()
	mockLocker.LockFunc.SetDefaultReturn(false, nil, nil)

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.RefDescriptionsFunc.SetDefaultReturn(map[string][]gitdomain.RefDescription{
		"b": {{IsDefaultBranch: true}},
	}, nil)

	updater := &updater{
		dbStore:         mockDBStore,
		locker:          mockLocker,
		gitserverClient: mockGitserverClient,
		operations:      NewOperations(mockDBStore, mockUploadSvc, &observation.TestContext),
	}

	if err := updater.Handle(context.Background()); err != nil {
		t.Fatalf("unexpected error updating commit graph: %s", err)
	}

	if len(mockUploadSvc.UpdateUploadsVisibleToCommitsFunc.History()) != 0 {
		t.Fatalf("unexpected calculate visible uploads call count. want=%d have=%d", 0, len(mockUploadSvc.UpdateUploadsVisibleToCommitsFunc.History()))
	}
}
