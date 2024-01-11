package service

import (
	"context"
	"io"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestMatchJsonWriter(t *testing.T) {
	mockStore := setupMockStore(t)

	matchJSONWriter, err := NewJSONWriter(context.Background(), mockStore, "1")
	require.NoError(t, err)

	testMatch1 := mkFileMatch(types.MinimalRepo{ID: 1, Name: "repo"}, "internal/search.go", 18, 27)
	testMatch2 := mkFileMatch(types.MinimalRepo{ID: 1, Name: "repo"}, "internal/service.go", 3, 7)

	err = matchJSONWriter.Write(testMatch1)
	require.NoError(t, err)

	err = matchJSONWriter.Write(testMatch2)
	require.NoError(t, err)

	err = matchJSONWriter.Close()
	require.NoError(t, err)

	iter, err := mockStore.List(context.Background(), "")
	require.NoError(t, err)
	uploadedFilesCount := 0
	for iter.Next() {
		uploadedFilesCount++
	}
	require.Equal(t, 1, uploadedFilesCount)

	blob, err := mockStore.Get(context.Background(), "1")
	require.NoError(t, err)

	blobBytes, err := io.ReadAll(blob)
	require.NoError(t, err)

	autogold.Expect(`{"type":"content","path":"internal/search.go","repositoryID":1,"repository":"repo","hunks":null,"chunkMatches":[{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":18,"column":0},"end":{"offset":0,"line":18,"column":0}}]},{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":27,"column":0},"end":{"offset":0,"line":27,"column":0}}]}]}
{"type":"content","path":"internal/service.go","repositoryID":1,"repository":"repo","hunks":null,"chunkMatches":[{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":3,"column":0},"end":{"offset":0,"line":3,"column":0}}]},{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":7,"column":0},"end":{"offset":0,"line":7,"column":0}}]}]}
`).Equal(t, string(blobBytes))
}

func TestBufferedWriter(t *testing.T) {
	mockStore := setupMockStore(t)

	blobUploader := &blobUploader{
		ctx:    context.Background(),
		store:  mockStore,
		prefix: "blob",
		shard:  1,
	}

	w := &bufferedWriter{
		maxSizeBytes: 8,
		w:            blobUploader,
	}

	n, err := w.Write([]byte("aaa\n")) // 4 bytes
	require.Equal(t, 4, n)
	require.NoError(t, err)

	n, err = w.Write([]byte("bbb\n"))
	require.Equal(t, 4, n)
	require.NoError(t, err)

	// We expect a new file to be created here because we have reached the max blob size.
	n, err = w.Write([]byte("ccc\n"))
	require.Equal(t, 4, n)
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)

	iter, err := mockStore.List(context.Background(), "")
	require.NoError(t, err)
	uploadedFilesCount := 0
	for iter.Next() {
		uploadedFilesCount++
	}
	require.Equal(t, 2, uploadedFilesCount)

	tc := []struct {
		wantKey  string
		wantBlob []byte
	}{
		{
			wantKey:  "blob",
			wantBlob: []byte("aaa\nbbb\n"),
		},
		{
			wantKey:  "blob-2",
			wantBlob: []byte("ccc\n"),
		},
	}

	for _, c := range tc {
		blob, err := mockStore.Get(context.Background(), c.wantKey)
		require.NoError(t, err)

		blobBytes, err := io.ReadAll(blob)
		require.NoError(t, err)

		require.Equal(t, c.wantBlob, blobBytes)
	}
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
