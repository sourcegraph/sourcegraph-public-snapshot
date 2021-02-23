package resolvers

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
	commitChecker := newCachedCommitChecker(mockGitserverClient)

	// known repository and commit
	mockDBStore.HasRepositoryFunc.SetDefaultReturn(true, nil)
	mockDBStore.HasCommitFunc.SetDefaultReturn(true, nil)

	// use commit graph in database
	mockDBStore.FindClosestDumpsFunc.SetDefaultReturn([]store.Dump{
		{ID: 50, RepositoryID: 42, Commit: "c0", Root: "s1/"},
		{ID: 51, RepositoryID: 42, Commit: "c1", Root: "s1/"}, // not in LSIF
		{ID: 52, RepositoryID: 42, Commit: "c2", Root: "s1/"},
		{ID: 53, RepositoryID: 42, Commit: "c3", Root: "s2/"}, // no file in root
		{ID: 54, RepositoryID: 42, Commit: "c4", Root: "s1/"}, // not in gitserver
	}, nil)

	mockLSIFStore.ExistsFunc.SetDefaultHook(func(ctx context.Context, bundleID int, path string) (bool, error) {
		return path == "main.go" && bundleID != 51, nil
	})
	mockGitserverClient.CommitExistsFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string) (bool, error) {
		return commit != "c4", nil
	})

	resolver := newResolver(mockDBStore, mockLSIFStore, mockGitserverClient, nil, nil, &observation.TestContext)
	dumps, err := resolver.findClosestDumps(context.Background(), commitChecker, 42, "deadbeef", "s1/main.go", true, "idx")
	if err != nil {
		t.Fatalf("unexpected error finding closest dumps: %s", err)
	}

	expected := []store.Dump{
		{ID: 50, RepositoryID: 42, Commit: "c0", Root: "s1/"},
		{ID: 52, RepositoryID: 42, Commit: "c2", Root: "s1/"},
	}
	if diff := cmp.Diff(expected, dumps); diff != "" {
		t.Errorf("unexpected dumps (-want +got):\n%s", diff)
	}
}

func TestFindClosestDumpsInfersClosestUploads(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	commitChecker := newCachedCommitChecker(mockGitserverClient)

	graph := gitserver.ParseCommitGraph([]string{
		"d",
		"c",
		"b d",
		"a b c",
	})

	// has repository, commit unknown but does exist
	mockDBStore.HasRepositoryFunc.SetDefaultReturn(true, nil)
	mockGitserverClient.CommitExistsFunc.SetDefaultReturn(true, nil)

	mockGitserverClient.CommitGraphFunc.SetDefaultReturn(graph, nil)
	mockDBStore.FindClosestDumpsFromGraphFragmentFunc.SetDefaultReturn([]store.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 51, Root: "s1/"},
		{ID: 52, Root: "s1/"},
		{ID: 53, Root: "s2/"},
	}, nil)
	mockLSIFStore.ExistsFunc.SetDefaultHook(func(ctx context.Context, bundleID int, path string) (bool, error) {
		if bundleID == 50 && path == "main.go" {
			return true, nil
		}
		if bundleID == 51 && path == "main.go" {
			return false, nil
		}
		if bundleID == 52 && path == "main.go" {
			return true, nil
		}
		if bundleID == 53 && path == "s1/main.go" {
			return false, nil
		}
		return false, nil
	})

	resolver := newResolver(mockDBStore, mockLSIFStore, mockGitserverClient, nil, nil, &observation.TestContext)
	dumps, err := resolver.findClosestDumps(context.Background(), commitChecker, 42, "deadbeef", "s1/main.go", true, "idx")
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

	if calls := mockDBStore.FindClosestDumpsFromGraphFragmentFunc.History(); len(calls) != 1 {
		t.Errorf("expected number of calls to store.FindClosestDumpsFromGraphFragmentFunc. want=%d have=%d", 1, len(calls))
	} else {
		expectedGraph := map[string][]string{
			"a": {"b", "c"},
			"b": {"d"},
			"c": {},
			"d": {},
		}
		if diff := cmp.Diff(expectedGraph, calls[0].Arg6.Graph()); diff != "" {
			t.Errorf("unexpected graph (-want +got):\n%s", diff)
		}
	}

	if value := len(mockDBStore.MarkRepositoryAsDirtyFunc.History()); value != 1 {
		t.Errorf("expected number of calls to store.MarkRepositoryAsDirty. want=%d have=%d", 1, value)
	}
}

func TestFindClosestDumpsDoesNotInferClosestUploadForUnknownRepository(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	commitChecker := newCachedCommitChecker(mockGitserverClient)

	resolver := newResolver(mockDBStore, mockLSIFStore, mockGitserverClient, nil, nil, &observation.TestContext)
	dumps, err := resolver.findClosestDumps(context.Background(), commitChecker, 42, "deadbeef", "s1/main.go", true, "idx")
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
