package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore/mocks"
	bundlemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore/mocks"
)

func TestFindClosestDumps(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockStoreHasRepository(t, mockStore, 42, true)
	setMockStoreHasCommit(t, mockStore, 42, testCommit, true)
	setMockStoreFindClosestDumps(t, mockStore, 42, testCommit, "s1/main.go", true, "idx", []store.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 51, Root: "s1/"},
		{ID: 52, Root: "s1/"},
		{ID: 53, Root: "s2/"},
	})
	setMultiMockBundleStoreExists(
		t,
		mockBundleStore,
		existsSpec{50, "main.go", true},
		existsSpec{51, "main.go", false},
		existsSpec{52, "main.go", true},
		existsSpec{53, "s1/main.go", false},
	)

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
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
}

func TestFindClosestDumpsInfersClosestUploads(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()

	graph := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {},
		"d": {},
	}

	setMockStoreHasRepository(t, mockStore, 42, true)
	setMockStoreHasCommit(t, mockStore, 42, testCommit, false)
	setMockGitserverCommitGraph(t, mockGitserverClient, 42, graph)
	setMockStoreFindClosestDumpsFromGraphFragment(t, mockStore, 42, testCommit, "s1/main.go", true, "idx", graph, []store.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 51, Root: "s1/"},
		{ID: 52, Root: "s1/"},
		{ID: 53, Root: "s2/"},
	})
	setMultiMockBundleStoreExists(
		t,
		mockBundleStore,
		existsSpec{50, "main.go", true},
		existsSpec{51, "main.go", false},
		existsSpec{52, "main.go", true},
		existsSpec{53, "s1/main.go", false},
	)

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
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

	if value := len(mockStore.MarkRepositoryAsDirtyFunc.History()); value != 1 {
		t.Errorf("expected number of calls to store.MarkRepositoryAsDirty. want=%d have=%d", 1, value)
	}
}

func TestFindClosestDumpsDoesNotInferClosestUploadForUnknownRepository(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockStoreHasRepository(t, mockStore, 42, false)
	setMockStoreHasCommit(t, mockStore, 42, testCommit, false)

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
	dumps, err := api.FindClosestDumps(context.Background(), 42, testCommit, "s1/main.go", true, "idx")
	if err != nil {
		t.Fatalf("unexpected error finding closest dumps: %s", err)
	}
	if len(dumps) != 0 {
		t.Errorf("unexpected number of dumps. want=%d have=%d", 0, len(dumps))
	}

	if value := len(mockStore.MarkRepositoryAsDirtyFunc.History()); value != 0 {
		t.Errorf("expected number of calls to store.MarkRepositoryAsDirty. want=%d have=%d", 0, value)
	}
}
