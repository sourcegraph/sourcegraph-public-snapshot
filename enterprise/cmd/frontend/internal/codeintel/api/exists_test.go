package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestFindClosestDumps(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockDBStoreHasRepository(t, mockDBStore, 42, true)
	setMockDBStoreHasCommit(t, mockDBStore, 42, testCommit, true)
	setMockDBStoreFindClosestDumps(t, mockDBStore, 42, testCommit, "s1/main.go", true, "idx", []store.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 51, Root: "s1/"},
		{ID: 52, Root: "s1/"},
		{ID: 53, Root: "s2/"},
	})
	setMultimockLSIFStoreExists(
		t,
		mockLSIFStore,
		existsSpec{50, "main.go", true},
		existsSpec{51, "main.go", false},
		existsSpec{52, "main.go", true},
		existsSpec{53, "s1/main.go", false},
	)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	graph := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {},
		"d": {},
	}

	setMockDBStoreHasRepository(t, mockDBStore, 42, true)
	setMockDBStoreHasCommit(t, mockDBStore, 42, testCommit, false)
	setMockGitserverCommitGraph(t, mockGitserverClient, 42, graph)
	setMockDBStoreFindClosestDumpsFromGraphFragment(t, mockDBStore, 42, testCommit, "s1/main.go", true, "idx", graph, []store.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 51, Root: "s1/"},
		{ID: 52, Root: "s1/"},
		{ID: 53, Root: "s2/"},
	})
	setMultimockLSIFStoreExists(
		t,
		mockLSIFStore,
		existsSpec{50, "main.go", true},
		existsSpec{51, "main.go", false},
		existsSpec{52, "main.go", true},
		existsSpec{53, "s1/main.go", false},
	)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
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

	if value := len(mockDBStore.MarkRepositoryAsDirtyFunc.History()); value != 1 {
		t.Errorf("expected number of calls to store.MarkRepositoryAsDirty. want=%d have=%d", 1, value)
	}
}

func TestFindClosestDumpsDoesNotInferClosestUploadForUnknownRepository(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockDBStoreHasRepository(t, mockDBStore, 42, false)
	setMockDBStoreHasCommit(t, mockDBStore, 42, testCommit, false)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
	dumps, err := api.FindClosestDumps(context.Background(), 42, testCommit, "s1/main.go", true, "idx")
	if err != nil {
		t.Fatalf("unexpected error finding closest dumps: %s", err)
	}
	if len(dumps) != 0 {
		t.Errorf("unexpected number of dumps. want=%d have=%d", 0, len(dumps))
	}

	if value := len(mockDBStore.MarkRepositoryAsDirtyFunc.History()); value != 0 {
		t.Errorf("expected number of calls to store.MarkRepositoryAsDirty. want=%d have=%d", 0, value)
	}
}
