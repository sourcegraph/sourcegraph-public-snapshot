package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
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
		{ID: 50, Root: "s1/", RepositoryID: 42, Commit: testCommit},
		{ID: 51, Root: "s1/", RepositoryID: 42, Commit: testCommit},
		{ID: 52, Root: "s1/", RepositoryID: 42, Commit: testCommit},
		{ID: 53, Root: "s2/", RepositoryID: 42, Commit: testCommit},
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
		{ID: 50, Root: "s1/", RepositoryID: 42, Commit: testCommit},
		{ID: 52, Root: "s1/", RepositoryID: 42, Commit: testCommit},
	}
	if diff := cmp.Diff(expected, dumps); diff != "" {
		t.Errorf("unexpected dumps (-want +got):\n%s", diff)
	}
}

func TestFindClosestDumpsInfersClosestUploads(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	graph := gitserver.ParseCommitGraph([]string{
		"d",
		"c",
		"b d",
		"a b c",
	})
	expectedGraph := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {},
		"d": {},
	}

	dumpsCommit, targetCommit := makeCommit(1), makeCommit(2)

	setMockDBStoreHasRepository(t, mockDBStore, 42, true)
	setMultiMockDBStoreHasCommit(t, mockDBStore, 42, []commitSpec{
		{exists: false, commit: targetCommit},
		{exists: true, commit: dumpsCommit},
	})
	setMockGitserverCommitGraph(t, mockGitserverClient, 42, graph)
	setMockDBStoreFindClosestDumpsFromGraphFragment(t, mockDBStore, 42, targetCommit, "s1/main.go", true, "idx", expectedGraph, []store.Dump{
		{ID: 50, Root: "s1/", RepositoryID: 42, Commit: dumpsCommit},
		{ID: 51, Root: "s1/", RepositoryID: 42, Commit: dumpsCommit},
		{ID: 52, Root: "s1/", RepositoryID: 42, Commit: dumpsCommit},
		{ID: 53, Root: "s2/", RepositoryID: 42, Commit: dumpsCommit},
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
	dumps, err := api.FindClosestDumps(context.Background(), 42, targetCommit, "s1/main.go", true, "idx")
	if err != nil {
		t.Fatalf("unexpected error finding closest dumps: %s", err)
	}

	expected := []store.Dump{
		{ID: 50, Root: "s1/", RepositoryID: 42, Commit: dumpsCommit},
		{ID: 52, Root: "s1/", RepositoryID: 42, Commit: dumpsCommit},
	}
	if diff := cmp.Diff(expected, dumps); diff != "" {
		t.Errorf("unexpected dumps (-want +got):\n%s", diff)
	}

	if value := len(mockDBStore.MarkRepositoryAsDirtyFunc.History()); value != 1 {
		t.Errorf("expected number of calls to store.MarkRepositoryAsDirty. want=%d have=%d", 1, value)
	}
}

func TestFindClosestDumpsDoesNotIncludeOrphanedCommits(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	graph := gitserver.ParseCommitGraph([]string{
		"d",
		"c",
		"b d",
		"a b c",
	})
	expectedGraph := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {},
		"d": {},
	}

	dumpsCommit, orphanedCommit, targetCommit := makeCommit(1), makeCommit(2), makeCommit(3)

	setMockDBStoreHasRepository(t, mockDBStore, 42, true)
	setMultiMockDBStoreHasCommit(t, mockDBStore, 42, []commitSpec{
		{commit: dumpsCommit, exists: true},
		{commit: targetCommit, exists: false},
		{commit: orphanedCommit, exists: false},
	})
	setMockGitserverCommitGraph(t, mockGitserverClient, 42, graph)
	setMockDBStoreFindClosestDumpsFromGraphFragment(t, mockDBStore, 42, targetCommit, "s1/main.go", true, "idx", expectedGraph, []store.Dump{
		{ID: 50, Root: "s1/", RepositoryID: 42, Commit: dumpsCommit},
		{ID: 51, Root: "s1/", RepositoryID: 42, Commit: dumpsCommit},
		{ID: 52, Root: "s1/", RepositoryID: 42, Commit: orphanedCommit},
		{ID: 53, Root: "s2/", RepositoryID: 42, Commit: dumpsCommit},
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
	dumps, err := api.FindClosestDumps(context.Background(), 42, targetCommit, "s1/main.go", true, "idx")
	if err != nil {
		t.Fatalf("unexpected error finding closest dumps: %s", err)
	}

	expected := []store.Dump{
		{ID: 50, Root: "s1/", RepositoryID: 42, Commit: dumpsCommit},
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
