package search

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// TestZipCacheDelete ensures that zip cache deletion is correctly hooked up to cache eviction.
func TestZipCacheDelete(t *testing.T) {
	// Set up a store.
	s := tmpStore(t)

	s.FetchTar = func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
		return emptyTar(t), nil
	}

	// Grab a zip.
	path, err := s.PrepareZip(context.Background(), "somerepo", "0123456789012345678901234567890123456789", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure it's there.
	_, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	// Load into zip cache.
	zf, err := s.zipCache.Get(path)
	if err != nil {
		t.Fatal(err)
	}
	zf.Close() // don't block eviction of this zipFile

	// Make sure it's there.
	if n := s.zipCache.count(); n != 1 {
		t.Fatalf("expected 1 item in cache, got %d", n)
	}

	// Evict from the store's disk cache.
	_, err = s.cache.Evict(0)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the zipFile is gone from the zip cache, too.
	if n := s.zipCache.count(); n != 0 {
		t.Fatalf("expected 0 items in cache, got %d", n)
	}

	// Make sure the file was successfully deleted on disk.
	_, err = os.Stat(path)
	if !os.IsNotExist(err) {
		t.Errorf("expected non-existence error, got %v", err)
	}
}
