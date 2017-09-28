package vfsutil

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestCachedFetch(t *testing.T) {
	cleanup := useEmptyArchiveCacheDir()
	defer cleanup()

	do := func() (Evicter, bool) {
		want := "foobar"
		calledFetcher := false
		f, err := cachedFetch(context.Background(), "component", "key", func(ctx context.Context) (io.ReadCloser, error) {
			calledFetcher = true
			return ioutil.NopCloser(bytes.NewReader([]byte(want))), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := ioutil.ReadAll(f.File)
		if err != nil {
			t.Fatal(err)
		}
		f.File.Close()
		if string(got) != want {
			t.Fatalf("did not return fetcher output. got %q, want %q", string(got), want)
		}
		return f, !calledFetcher
	}

	// Cache should be empty
	_, usedCache := do()
	if usedCache {
		t.Fatal("Expected fetcher to be called on empty cache")
	}

	// Redo, now we should use the cache
	f, usedCache := do()
	if !usedCache {
		t.Fatal("Expected fetcher to not be called when cached")
	}

	// Evict, then we should not use the cache
	f.Evict()
	_, usedCache = do()
	if usedCache {
		t.Fatal("Item was not properly evicted")
	}
}

func useEmptyArchiveCacheDir() func() {
	d, err := ioutil.TempDir("", "vfsutil_test")
	if err != nil {
		panic(err)
	}
	orig := ArchiveCacheDir
	ArchiveCacheDir = d
	return func() {
		os.RemoveAll(d)
		ArchiveCacheDir = orig
	}
}
