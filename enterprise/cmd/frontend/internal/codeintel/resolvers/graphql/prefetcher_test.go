package graphql

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	resolvermocks "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
)

func TestPrefetcherUploads(t *testing.T) {
	mockResolver := resolvermocks.NewMockResolver()
	prefetcher := NewPrefetcher(mockResolver)

	uploads := map[int]dbstore.Upload{
		1: {ID: 1},
		2: {ID: 2},
		3: {ID: 3},
		4: {ID: 4},
		5: {ID: 5},
	}

	mockResolver.GetUploadsByIDsFunc.SetDefaultHook(func(_ context.Context, ids ...int) ([]dbstore.Upload, error) {
		matching := make([]dbstore.Upload, 0, len(ids))
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
	} else if callCount := len(mockResolver.GetUploadsByIDsFunc.History()); callCount != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, callCount)
	}

	// Re-fetch cached
	if upload, exists, err := prefetcher.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error fetching upload: %s", err)
	} else if !exists {
		t.Fatalf("expected upload to exist")
	} else if diff := cmp.Diff(uploads[1], upload); diff != "" {
		t.Fatalf("unexpected upload (-want +got):\n%s", diff)
	} else if callCount := len(mockResolver.GetUploadsByIDsFunc.History()); callCount != 1 {
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
	} else if callCount := len(mockResolver.GetUploadsByIDsFunc.History()); callCount != 2 {
		t.Fatalf("unexpected call count. want=%d have=%d", 2, callCount)
	}

	// Cached from earlier
	if upload, exists, err := prefetcher.GetUploadByID(context.Background(), 4); err != nil {
		t.Fatalf("unexpected error fetching upload: %s", err)
	} else if !exists {
		t.Fatalf("expected upload to exist")
	} else if diff := cmp.Diff(uploads[4], upload); diff != "" {
		t.Fatalf("unexpected upload (-want +got):\n%s", diff)
	} else if callCount := len(mockResolver.GetUploadsByIDsFunc.History()); callCount != 2 {
		t.Fatalf("unexpected call count. want=%d have=%d", 2, callCount)
	}
}

func TestPrefetcherIndexes(t *testing.T) {
	mockResolver := resolvermocks.NewMockResolver()
	prefetcher := NewPrefetcher(mockResolver)

	indexes := map[int]dbstore.Index{
		1: {ID: 1},
		2: {ID: 2},
		3: {ID: 3},
		4: {ID: 4},
		5: {ID: 5},
	}

	mockResolver.GetIndexesByIDsFunc.SetDefaultHook(func(_ context.Context, ids ...int) ([]dbstore.Index, error) {
		matching := make([]dbstore.Index, 0, len(ids))
		for _, id := range ids {
			matching = append(matching, indexes[id])
		}

		return matching, nil
	})

	// Bare fetch
	if index, exists, err := prefetcher.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error fetching index: %s", err)
	} else if !exists {
		t.Fatalf("expected index to exist")
	} else if diff := cmp.Diff(indexes[1], index); diff != "" {
		t.Fatalf("unexpected index (-want +got):\n%s", diff)
	} else if callCount := len(mockResolver.GetIndexesByIDsFunc.History()); callCount != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, callCount)
	}

	// Re-fetch cached
	if index, exists, err := prefetcher.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error fetching index: %s", err)
	} else if !exists {
		t.Fatalf("expected index to exist")
	} else if diff := cmp.Diff(indexes[1], index); diff != "" {
		t.Fatalf("unexpected index (-want +got):\n%s", diff)
	} else if callCount := len(mockResolver.GetIndexesByIDsFunc.History()); callCount != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, callCount)
	}

	// Fetch batch
	prefetcher.MarkIndex(2)
	prefetcher.MarkIndex(3)
	prefetcher.MarkIndex(4)
	prefetcher.MarkIndex(6) // unknown id

	if index, exists, err := prefetcher.GetIndexByID(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error fetching index: %s", err)
	} else if !exists {
		t.Fatalf("expected index to exist")
	} else if diff := cmp.Diff(indexes[2], index); diff != "" {
		t.Fatalf("unexpected index (-want +got):\n%s", diff)
	} else if callCount := len(mockResolver.GetIndexesByIDsFunc.History()); callCount != 2 {
		t.Fatalf("unexpected call count. want=%d have=%d", 2, callCount)
	}

	// Cached from earlier
	if index, exists, err := prefetcher.GetIndexByID(context.Background(), 4); err != nil {
		t.Fatalf("unexpected error fetching index: %s", err)
	} else if !exists {
		t.Fatalf("expected index to exist")
	} else if diff := cmp.Diff(indexes[4], index); diff != "" {
		t.Fatalf("unexpected index (-want +got):\n%s", diff)
	} else if callCount := len(mockResolver.GetIndexesByIDsFunc.History()); callCount != 2 {
		t.Fatalf("unexpected call count. want=%d have=%d", 2, callCount)
	}
}
