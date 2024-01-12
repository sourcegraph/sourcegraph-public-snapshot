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

	autogold.Expect(`[{"type":"content","path":"internal/search.go","repositoryID":1,"repository":"repo","hunks":null,"chunkMatches":[{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":18,"column":0},"end":{"offset":0,"line":18,"column":0}}]},{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":27,"column":0},"end":{"offset":0,"line":27,"column":0}}]}]},{"type":"content","path":"internal/service.go","repositoryID":1,"repository":"repo","hunks":null,"chunkMatches":[{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":3,"column":0},"end":{"offset":0,"line":3,"column":0}}]},{"content":"","contentStart":{"offset":0,"line":0,"column":0},"ranges":[{"start":{"offset":0,"line":7,"column":0},"end":{"offset":0,"line":7,"column":0}}]}]}]`).Equal(t, string(blobBytes))
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
