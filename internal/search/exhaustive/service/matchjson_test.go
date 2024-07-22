package service

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/object/mocks"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

func TestMatchJsonWriter(t *testing.T) {
	mockStore := setupMockStore(t)

	matchJSONWriter, err := NewJSONWriter(context.Background(), mockStore, "dummy_prefix")
	require.NoError(t, err)

	testMatch1 := mkFileMatch(types.MinimalRepo{ID: 1, Name: "repo"}, "internal/search.go", 18, 27)
	testMatch2 := mkFileMatch(types.MinimalRepo{ID: 1, Name: "repo"}, "internal/service.go", 3, 7)

	err = matchJSONWriter.Write(testMatch1)
	require.NoError(t, err)

	err = matchJSONWriter.Write(testMatch2)
	require.NoError(t, err)

	err = matchJSONWriter.Flush()
	require.NoError(t, err)

	iter, err := mockStore.List(context.Background(), "")
	require.NoError(t, err)
	uploadedFilesCount := 0
	for iter.Next() {
		uploadedFilesCount++
	}
	require.Equal(t, 1, uploadedFilesCount)

	blob, err := mockStore.Get(context.Background(), "dummy_prefix")
	require.NoError(t, err)

	blobBytes, err := io.ReadAll(blob)
	require.NoError(t, err)

	autogold.Expect(`{"type":"content","path":"internal/search.go","repositoryID":1,"repository":"repo","hunks":null,"chunkMatches":[{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":18,"column":0},"end":{"offset":0,"line":18,"column":0}}]},{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":27,"column":0},"end":{"offset":0,"line":27,"column":0}}]}],"language":"Go"}
{"type":"content","path":"internal/service.go","repositoryID":1,"repository":"repo","hunks":null,"chunkMatches":[{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":3,"column":0},"end":{"offset":0,"line":3,"column":0}}]},{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":7,"column":0},"end":{"offset":0,"line":7,"column":0}}]}],"language":"Go"}
`).Equal(t, string(blobBytes))
}

func mkFileMatch(repo types.MinimalRepo, path string, lineNumbers ...int) *result.FileMatch {
	var hms result.ChunkMatches
	for _, n := range lineNumbers {
		hms = append(hms, result.ChunkMatch{
			Ranges: []result.Range{{
				Start: result.Location{Line: n},
				End:   result.Location{Line: n},
			}},
		})
	}

	return &result.FileMatch{
		File: result.File{
			Path: path,
			Repo: repo,
		},
		ChunkMatches: hms,
	}
}

func TestBufferedWriter(t *testing.T) {
	mockStore := setupMockStore(t)

	uploader := blobUploader{
		ctx:    context.Background(),
		store:  mockStore,
		prefix: "blob",
		shard:  1,
	}

	w := newBufferedWriter(24, uploader.write)

	testData := func(val string) any {
		return struct{ Key string }{Key: val}
	}

	err := w.Append(testData("a")) // {"Key":"a"}\n 12 bytes
	require.NoError(t, err)
	err = w.Append(testData("b"))
	require.NoError(t, err)

	// We expect a new file to be created here because we have reached the max blob size.
	err = w.Append(testData("c"))
	require.NoError(t, err)

	err = w.Flush()
	require.NoError(t, err)

	wantFiles := 2
	iter, err := mockStore.List(context.Background(), "")
	require.NoError(t, err)
	haveFiles := 0
	for iter.Next() {
		haveFiles++
	}
	require.Equal(t, wantFiles, haveFiles)

	tc := []struct {
		wantKey  string
		wantBlob string
	}{
		{
			wantKey:  "blob",
			wantBlob: "{\"Key\":\"a\"}\n{\"Key\":\"b\"}\n",
		},
		{
			wantKey:  "blob-2",
			wantBlob: "{\"Key\":\"c\"}\n",
		},
	}

	for _, c := range tc {
		blob, err := mockStore.Get(context.Background(), c.wantKey)
		require.NoError(t, err)

		blobBytes, err := io.ReadAll(blob)
		require.NoError(t, err)

		require.Equal(t, c.wantBlob, string(blobBytes))
	}
}

func TestNoUploadIfNotData(t *testing.T) {
	mockStore := setupMockStore(t)

	w, err := NewJSONWriter(context.Background(), mockStore, "dummy_prefix")
	require.NoError(t, err)

	// No data written, so no upload should happen.
	err = w.Flush()
	require.NoError(t, err)
	iter, err := mockStore.List(context.Background(), "")
	require.NoError(t, err)

	for iter.Next() {
		t.Fatal("should not have uploaded anything")
	}
}

func setupMockStore(t *testing.T) *mocks.MockStorage {
	t.Helper()

	bucket := make(map[string][]byte)

	mockStore := mocks.NewMockStorage()
	mockStore.UploadFunc.SetDefaultHook(func(ctx context.Context, key string, r io.Reader) (int64, error) {
		b, err := io.ReadAll(r)
		if err != nil {
			return 0, err
		}

		bucket[key] = b

		return int64(len(b)), nil
	})

	mockStore.ListFunc.SetDefaultHook(func(ctx context.Context, prefix string) (*iterator.Iterator[string], error) {
		var keys []string
		for k := range bucket {
			if strings.HasPrefix(k, prefix) {
				keys = append(keys, k)
			}
		}
		return iterator.From(keys), nil
	})

	mockStore.GetFunc.SetDefaultHook(func(ctx context.Context, key string) (io.ReadCloser, error) {
		if b, ok := bucket[key]; ok {
			return io.NopCloser(bytes.NewReader(b)), nil
		}
		return nil, errors.New("key not found")
	})

	return mockStore
}
