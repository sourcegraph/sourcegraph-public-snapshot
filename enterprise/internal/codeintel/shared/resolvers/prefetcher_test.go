package sharedresolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

func TestPrefetcherUploads(t *testing.T) {
	mockAutoIndexingResolver := NewMockAutoIndexingService()
	mockUploadResolver := NewMockUploadsService()
	prefetcher := NewPrefetcher(mockAutoIndexingResolver, mockUploadResolver)

	uploads := map[int]types.Upload{
		1: {ID: 1},
		2: {ID: 2},
		3: {ID: 3},
		4: {ID: 4},
		5: {ID: 5},
	}

	mockUploadResolver.GetUploadsByIDsFunc.SetDefaultHook(func(_ context.Context, ids ...int) ([]types.Upload, error) {
		matching := make([]types.Upload, 0, len(ids))
		for _, id := range ids {
			matching = append(matching, uploads[id])
		}

		return matching, nil
	})

	// Bare fetch
	if upload, exists, err := prefetcher.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error fetching upload: %s", err)
	} else if !exists {
		t.Fatalf("expected upload to exist")
	} else if diff := cmp.Diff(uploads[1], upload); diff != "" {
		t.Fatalf("unexpected upload (-want +got):\n%s", diff)
	} else if callCount := len(mockUploadResolver.GetUploadsByIDsFunc.History()); callCount != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, callCount)
	}

	// Re-fetch cached
	if upload, exists, err := prefetcher.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error fetching upload: %s", err)
	} else if !exists {
		t.Fatalf("expected upload to exist")
	} else if diff := cmp.Diff(uploads[1], upload); diff != "" {
		t.Fatalf("unexpected upload (-want +got):\n%s", diff)
	} else if callCount := len(mockUploadResolver.GetUploadsByIDsFunc.History()); callCount != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, callCount)
	}

	// Fetch batch
	prefetcher.MarkUpload(2)
	prefetcher.MarkUpload(3)
	prefetcher.MarkUpload(4)
	prefetcher.MarkUpload(6) // unknown id

	if upload, exists, err := prefetcher.GetUploadByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error fetching upload: %s", err)
	} else if !exists {
		t.Fatalf("expected upload to exist")
	} else if diff := cmp.Diff(uploads[2], upload); diff != "" {
		t.Fatalf("unexpected upload (-want +got):\n%s", diff)
	} else if callCount := len(mockUploadResolver.GetUploadsByIDsFunc.History()); callCount != 2 {
		t.Fatalf("unexpected call count. want=%d have=%d", 2, callCount)
	}

	// Cached from earlier
	if upload, exists, err := prefetcher.GetUploadByID(context.Background(), 4); err != nil {
		t.Fatalf("unexpected error fetching upload: %s", err)
	} else if !exists {
		t.Fatalf("expected upload to exist")
	} else if diff := cmp.Diff(uploads[4], upload); diff != "" {
		t.Fatalf("unexpected upload (-want +got):\n%s", diff)
	} else if callCount := len(mockUploadResolver.GetUploadsByIDsFunc.History()); callCount != 2 {
		t.Fatalf("unexpected call count. want=%d have=%d", 2, callCount)
	}
}

func TestPrefetcherIndexes(t *testing.T) {
	indexes := map[int]types.Index{
		1: {ID: 1},
		2: {ID: 2},
		3: {ID: 3},
		4: {ID: 4},
		5: {ID: 5},
	}

	mockAutoIndexingResolver := NewMockAutoIndexingService()
	mockAutoIndexingResolver.GetIndexesByIDsFunc.SetDefaultHook(func(_ context.Context, ids ...int) ([]types.Index, error) {
		matching := make([]types.Index, 0, len(ids))
		for _, id := range ids {
			matching = append(matching, indexes[id])
		}

		return matching, nil
	})
	mockUploadResolver := NewMockUploadsService()
	prefetcher := NewPrefetcher(mockAutoIndexingResolver, mockUploadResolver)

	// We do a conversion inside the function that I cannot reproduct inside the mock.
	expectedIndex := types.Index{ID: 1}

	// Bare fetch
	if index, exists, err := prefetcher.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error fetching index: %s", err)
	} else if !exists {
		t.Fatalf("expected index to exist")
	} else if diff := cmp.Diff(expectedIndex, index); diff != "" {
		t.Fatalf("unexpected index (-want +got):\n%s", diff)
	} else if callCount := len(mockAutoIndexingResolver.GetIndexesByIDsFunc.History()); callCount != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, callCount)
	}

	// Re-fetch cached
	if index, exists, err := prefetcher.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error fetching index: %s", err)
	} else if !exists {
		t.Fatalf("expected index to exist")
	} else if diff := cmp.Diff(expectedIndex, index); diff != "" {
		t.Fatalf("unexpected index (-want +got):\n%s", diff)
	} else if callCount := len(mockAutoIndexingResolver.GetIndexesByIDsFunc.History()); callCount != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, callCount)
	}

	// Fetch batch
	prefetcher.MarkIndex(2)
	prefetcher.MarkIndex(3)
	prefetcher.MarkIndex(4)
	prefetcher.MarkIndex(6) // unknown id

	mockAutoIndexingResolver.GetIndexesByIDsFunc.SetDefaultHook(func(_ context.Context, ids ...int) ([]types.Index, error) {
		matching := make([]types.Index, 0, len(ids))
		for _, id := range ids {
			matching = append(matching, indexes[id])
		}

		return matching, nil
	})

	expectedIndex.ID = 2
	if index, exists, err := prefetcher.GetIndexByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error fetching index: %s", err)
	} else if !exists {
		t.Fatalf("expected index to exist")
	} else if diff := cmp.Diff(expectedIndex, index); diff != "" {
		t.Fatalf("unexpected index (-want +got):\n%s", diff)
	} else if callCount := len(mockAutoIndexingResolver.GetIndexesByIDsFunc.History()); callCount != 2 {
		t.Fatalf("unexpected call count. want=%d have=%d", 2, callCount)
	}

	// Cached from earlier
	expectedIndex.ID = 4
	if index, exists, err := prefetcher.GetIndexByID(context.Background(), 4); err != nil {
		t.Fatalf("unexpected error fetching index: %s", err)
	} else if !exists {
		t.Fatalf("expected index to exist")
	} else if diff := cmp.Diff(expectedIndex, index); diff != "" {
		t.Fatalf("unexpected index (-want +got):\n%s", diff)
	} else if callCount := len(mockAutoIndexingResolver.GetIndexesByIDsFunc.History()); callCount != 2 {
		t.Fatalf("unexpected call count. want=%d have=%d", 2, callCount)
	}
}
