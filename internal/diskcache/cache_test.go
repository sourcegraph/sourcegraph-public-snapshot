package diskcache

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestOpen(t *testing.T) {
	dir := t.TempDir()

	store := &store{
		dir:       dir,
		component: "test",
		observe:   newOperations(&observation.TestContext, "test"),
	}

	do := func() (*File, bool) {
		want := "foobar"
		calledFetcher := false
		f, err := store.Open(context.Background(), []string{"key"}, func(ctx context.Context) (io.ReadCloser, error) {
			calledFetcher = true
			return io.NopCloser(bytes.NewReader([]byte(want))), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(f.File)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
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
	os.Remove(f.Path)
	_, usedCache = do()
	if usedCache {
		t.Fatal("Item was not properly evicted")
	}
}

func TestMultiKeyEviction(t *testing.T) {
	dir := t.TempDir()

	store := &store{
		dir:       dir,
		component: "test",
		observe:   newOperations(&observation.TestContext, "test"),
	}

	f, err := store.Open(context.Background(), []string{"key1", "key2"}, func(ctx context.Context) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader([]byte("blah"))), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	stats, err := store.Evict(0)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Evicted != 1 {
		t.Fatal("Expected to evict 1 item, evicted", stats.Evicted)
	}
}
