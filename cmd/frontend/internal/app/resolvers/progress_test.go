package resolvers

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestGetProgress(t *testing.T) {
	testCases := []struct {
		status                           repo.EmbedRepoStats
		expectedProcessed, expectedTotal *int32
		expectedProgress                 float64
	}{
		{
			repo.EmbedRepoStats{
				CodeIndexStats: repo.EmbedFilesStats{
					FilesScheduled: 10,
					FilesEmbedded:  5,
					FilesSkipped:   map[string]int{"small": 2, "large": 3},
				},
				TextIndexStats: repo.EmbedFilesStats{
					FilesScheduled: 10,
					FilesEmbedded:  5,
					FilesSkipped:   map[string]int{"small": 1, "large": 2},
				},
			},
			pointers.Ptr(int32(5 + 2 + 3 + 5 + 1 + 2)),
			pointers.Ptr(int32(10 + 10)),
			0.9,
		},
		{
			repo.EmbedRepoStats{
				CodeIndexStats: repo.EmbedFilesStats{
					FilesScheduled: 10,
					FilesEmbedded:  10,
				},
				TextIndexStats: repo.EmbedFilesStats{
					FilesScheduled: 10,
					FilesEmbedded:  10,
				},
			},
			pointers.Ptr(int32(10 + 10)),
			pointers.Ptr(int32(10 + 10)),
			1.0,
		},
	}

	for _, tc := range testCases {
		processed, total, progress := getProgress(tc.status)
		if *processed != *tc.expectedProcessed || *total != *tc.expectedTotal || progress != tc.expectedProgress {
			t.Errorf("Expected processed %d, total %d and progress %f but got %d, %d and %f", *tc.expectedProcessed, *tc.expectedTotal, tc.expectedProgress, *processed, *total, progress)
		}
	}
}
