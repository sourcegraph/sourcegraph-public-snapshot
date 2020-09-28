package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	bundlemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client/mocks"
	commitmocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/commits/mocks"
	gitservermocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver/mocks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
)

func TestFindClosestDumps(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient1 := bundlemocks.NewMockBundleClient()
	mockBundleClient2 := bundlemocks.NewMockBundleClient()
	mockBundleClient3 := bundlemocks.NewMockBundleClient()
	mockBundleClient4 := bundlemocks.NewMockBundleClient()
	mockGitserverClient := gitservermocks.NewMockClient()
	mockCommitUpdater := commitmocks.NewMockUpdater()

	setMockStoreHasRepository(t, mockStore, 42, true)
	setMockStoreFindClosestDumps(t, mockStore, 42, testCommit, "s1/main.go", true, "idx", []store.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 51, Root: "s1/"},
		{ID: 52, Root: "s1/"},
		{ID: 53, Root: "s2/"},
	})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{
		50: mockBundleClient1,
		51: mockBundleClient2,
		52: mockBundleClient3,
		53: mockBundleClient4,
	})
	setMockBundleClientExists(t, mockBundleClient1, "main.go", true)
	setMockBundleClientExists(t, mockBundleClient2, "main.go", false)
	setMockBundleClientExists(t, mockBundleClient3, "main.go", true)
	setMockBundleClientExists(t, mockBundleClient4, "s1/main.go", false)

	api := testAPI(mockStore, mockBundleManagerClient, mockGitserverClient, mockCommitUpdater)
	dumps, err := api.FindClosestDumps(context.Background(), 42, testCommit, "s1/main.go", true, "idx")
	if err != nil {
		t.Fatalf("unexpected error finding closest dumps: %s", err)
	}

	expected := []store.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 52, Root: "s1/"},
	}
	if diff := cmp.Diff(expected, dumps); diff != "" {
		t.Errorf("unexpected dumps (-want +got):\n%s", diff)
	}

	if len(mockCommitUpdater.UpdateFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdateFunc calls. want=%d have=%d", 1, len(mockCommitUpdater.UpdateFunc.History()))
	} else if mockCommitUpdater.UpdateFunc.History()[0].Arg1 != 42 {
		t.Errorf("unexpected value for repository id. want=%d have=%d", 42, mockCommitUpdater.UpdateFunc.History()[0].Arg1)
	}
}

func TestFindClosestSkipsCommitGraphUpdateIfCommitIsKnown(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()
	mockGitserverClient := gitservermocks.NewMockClient()
	mockCommitUpdater := commitmocks.NewMockUpdater()

	setMockStoreHasCommit(t, mockStore, 42, testCommit, true)
	setMockStoreFindClosestDumps(t, mockStore, 42, testCommit, "main.go", true, "idx", []store.Dump{
		{ID: 50, Root: ""},
	})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{
		50: mockBundleClient,
	})
	setMockBundleClientExists(t, mockBundleClient, "main.go", true)

	api := New(mockStore, mockBundleManagerClient, mockGitserverClient, mockCommitUpdater)
	dumps, err := api.FindClosestDumps(context.Background(), 42, testCommit, "main.go", true, "idx")
	if err != nil {
		t.Fatalf("unexpected error finding closest dumps: %s", err)
	}

	expected := []store.Dump{{ID: 50, Root: ""}}
	if diff := cmp.Diff(expected, dumps); diff != "" {
		t.Errorf("unexpected dumps (-want +got):\n%s", diff)
	}

	if len(mockCommitUpdater.UpdateFunc.History()) != 0 {
		t.Errorf("expected commitUpdater.Update not to be called")
	}
}

func TestFindClosestSkipsCommitGraphUpdateIfRepositoryIsUnknown(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockGitserverClient := gitservermocks.NewMockClient()
	mockCommitUpdater := commitmocks.NewMockUpdater()

	setMockStoreHasRepository(t, mockStore, 42, false)

	api := New(mockStore, mockBundleManagerClient, mockGitserverClient, mockCommitUpdater)
	dumps, err := api.FindClosestDumps(context.Background(), 42, testCommit, "main.go", true, "idx")
	if err != nil {
		t.Fatalf("unexpected error finding closest dumps: %s", err)
	}
	if len(dumps) != 0 {
		t.Errorf("unexpected dumps")
	}

	if len(mockStore.FindClosestDumpsFunc.History()) != 0 {
		t.Errorf("expected store.FindClosestDumps not to be called")
	}
	if len(mockCommitUpdater.UpdateFunc.History()) != 0 {
		t.Errorf("expected commitUpdater.Update not to be called")
	}
}
