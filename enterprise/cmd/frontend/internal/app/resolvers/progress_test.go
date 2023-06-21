package resolvers

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
)

func TestGetProgress(t *testing.T) {
	testCases := []struct {
		status   repo.EmbedRepoStats
		expected float64
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
			1.0,
		},
	}

	for _, tc := range testCases {
		progress := getProgress(tc.status)
		if progress != tc.expected {
			t.Errorf("Expected progress %f but got %f", tc.expected, progress)
		}
	}
}
